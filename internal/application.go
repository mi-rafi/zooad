package internal

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	//"github.com/mi-raf/zooad/internal"
	//"github.com/mi-raf/zooad/internal/service"
)

func (a *Application) init() error {
	if a.TerminationTimeout == 0 {
		a.TerminationTimeout = defaultTerminationTimeout
	}
	if a.InitializationTimeout == 0 {
		a.InitializationTimeout = defaultInitializationTimeout
	}
	a.holdOn = make(chan struct{})
	a.done = make(chan struct{})
	if a.Resources != nil {
		ctx, cansel := context.WithTimeout(a, a.InitializationTimeout)
		defer cansel()
		return a.Resources.Init(ctx)
	}
	return nil
}

type (
	Resources interface {
		Init(context.Context) error  // чтобы инициализировать
		Watch(context.Context) error // чтобы наблюдать
		Stop()                       // остановить наблюдение
		Release() error              // освободить ресурсы
	}

	Application struct {
		//service.ServiceKeeper
		// это будет выполняться основным потоком совсем хз как это работает
		MainFunc func(ctx context.Context, holdOn <-chan struct{}) error

		Resources             Resources
		TerminationTimeout    time.Duration
		InitializationTimeout time.Duration

		//пока тоже хз мьютекс и чан о_О
		appState int32
		err      error
		mux      sync.Mutex
		done     chan struct{}
		holdOn   chan struct{}
	}

	AppContext struct{}
)

const (
	appStateInit int32 = iota
	appStateRunning
	appStateHalt
	appStateShutdown
)

const (
	defaultTerminationTimeout    = time.Second
	defaultInitializationTimeout = time.Second * 15
)

func (a *Application) checkState(old, new int32) bool {
	return atomic.CompareAndSwapInt32(&a.appState, old, new)
}

func (a *Application) Run() error {
	if a.MainFunc == nil {
		return ErrMainOmitted
	}
	if a.checkState(appStateInit, appStateRunning) {
		if err := a.init(); err != nil {
			a.err = err
			a.appState = appStateShutdown
			return err
		}
		// с помощью servicesRunning мы синхронизируем жизненный цикл ресурсов
		// с жизненным циклом приложения
		var servicesRunning = make(chan struct{})
		if a.Resources != nil {
			go a.watchResources(servicesRunning)
			// go func() {
			// 	defer close(servicesRunning) // вот сигнал о том, что Watch остановлено
			// 	// Shutdown просто остновит a.run(sig), это мы потом увидим
			// 	defer a.Shutdown()
			// 	if err := a.Resources.Watch(context.TODO()); err != nil {
			// 		a.setError(err)
			// 	}

			// }()
		}
		//Это кто такой ??
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		// запускаем основной поток выполнения
		if err := a.run(sig); err != nil {
			a.setError(err)
		}
		// в этом месте программа должна завершиться
		if a.Resources != nil {
			a.Resources.Stop() // посылаем сигнал ресурсам
			select {
			case <-servicesRunning: // ожидаем завершения Watch
			case <-time.After(a.TerminationTimeout):
			}
			if err := a.Resources.Release(); err != nil { // освобождаем ресурсы
				a.setError(err)
			}

		}
		return a.getError()
	}
	return ErrWrongState
}

func (a *Application) Shutdown() {
	a.Halt()
	if a.checkState(appStateHalt, appStateShutdown) {
		close(a.done)
	}
}

func (a *Application) setError(err error) {
	if err == nil {
		return
	}
	a.mux.Lock()
	if a.err == nil {
		a.err = err
	}
	a.mux.Unlock()
	a.Shutdown()
}

func (a *Application) run(sig <-chan os.Signal) error {
	defer a.Shutdown()               // при выходе просто установит поле state в значение appStateShutdown
	var errRun = make(chan error, 1) // канал для сигнала от основного потока
	go func() {
		defer close(errRun)
		// halt для основного потока - это сигнал о завершении работы
		if err := a.MainFunc(a, a.holdOn); err != nil {
			errRun <- err
		}
	}()
	var errHld = make(chan error, 1)
	go func() {
		defer close(errHld)
		select {
		// ожидаем сигнала операционной системы
		case <-sig:
			a.Halt() // вызов этой процедуры просто закроет канал a.halt
			// это и будет наш Graceful Shutdown воркфлоу
			// нам нужно дождаться завершения основного потока или выйти по таймауту
			select {
			case <-time.After(a.TerminationTimeout):
				// это выход по таймауту
				errHld <- ErrTermTimeout
			case <-a.done: // a.Shutdown закрывает этот канал
				// ok
			}
		case <-a.done: // a.Shutdown закрывает этот канал
			// сюда попадем, если завершение работы произошло без участия ОС
		}
	}()
	// на этом месте выполнение процедуры будет блокировано
	// пока не произойдет одно из следующих событий
	select {
	// получим ошибку от основного потока выполнения или закроется канал errRun
	case err, ok := <-errRun:
		if ok && err != nil {
			return err
		}
	// получим ошибку от рутины, слушающей сигналы ОС или закроется ее канал
	case err, ok := <-errHld:
		if ok && err != nil {
			return err
		}
		// это жесткий путь - кто-то вызвал процедуру Shutdown()
	case <-a.done:
		// shutdown
	}
	return nil
}

func (a *Application) Halt() {
	if a.checkState(appStateRunning, appStateHalt) {
		close(a.holdOn)
	}
}

func (a *Application) getError() error {
	var err error
	a.mux.Lock()
	err = a.err
	a.mux.Unlock()
	return err
}

func (a *Application) Deadline() (deadline time.Time, ok bool) {
	return time.Time{}, false
}

func (a *Application) Done() <-chan struct{} {
	return a.done
}

func (a *Application) Err() error {
	if err := a.getError(); err != nil {
		return err
	}

	if atomic.LoadInt32(&a.appState) == appStateShutdown {
		return ErrShutdown
	}
	return nil
}

func (a *Application) Value(key interface{}) interface{} {
	var AppContext = AppContext{}
	if key == AppContext {
		return a
	}
	return nil
}

func (a *Application) watchResources(servicesRunning chan<- struct{}) {
	defer close(servicesRunning)
	defer a.Shutdown()
	a.setError(a.Resources.Watch(a))
}
