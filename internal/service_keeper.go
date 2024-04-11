package internal

import (
	"context"
	//"sync"
	"sync/atomic"
	"time"

	"github.com/mi-raf/zooad/internal/service"
)

const (
	srvStateInit int32 = iota
	srvStateReady
	srvStateRunning
	srvStateShutdown
	srvStateOff
)

const (
	defaultPingPeriod      = time.Second * 5
	defaultPingTimeout     = time.Millisecond * 1500
	defaultShutdownTimeout = time.Millisecond * 15000
)

func (s *ServiceKeeper) Init(ctx context.Context) error {
	if !s.checkState(srvStateInit, srvStateReady) {
		return ErrWrongState
	}
	if err := s.initAllServices(ctx); err != nil {
		return err
	}
	s.stop = make(chan struct{})
	if s.PingPeriod == 0 {
		s.PingPeriod = defaultPingPeriod
	}
	if s.PingTimeout == 0 {
		s.PingTimeout = defaultPingTimeout
	}
	if s.ShutdownTimeout == 0 {
		s.ShutdownTimeout = defaultShutdownTimeout
	}
	return nil
}

type ServiceKeeper struct {
	Services        []service.Service
	state           int32
	PingPeriod      time.Duration
	PingTimeout     time.Duration
	ShutdownTimeout time.Duration
	stop            chan struct{}
}

// не поняла зачем давать имя возращающей переменной
func (s *ServiceKeeper) initAllServices(ctx context.Context) (initError error) {
	initCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	var p paralleRun
	for i := range s.Services {
		p.do(initCtx, s.Services[i].Init)
		// if err := s.Services[i].Init(ctx); err != nil {
		// 	return err
		// }
	}
	return p.wait()
}

// если поменял значение
func (s *ServiceKeeper) checkState(old, new int32) bool {
	return atomic.CompareAndSwapInt32(&s.state, old, new)
}

// Теперь немного по поводу ServiceKeeper и его метода Watch. Вызов Watch должен быть блокирующий,
// ведь в нашем коде Application мы вызываем его только раз, и после его выполнения происходит немедленное
// завершение работы через вызов Shutdown. Что требуется от реализации этого метода:
// С некоторой периодичностью выполнять Ping ресурсов, которые зарегистрированы внутри ServiceKeeper.
// Прекращать циклическое выполнение Ping при обнаружении критической ошибки и возвращать error.
// Прекращать циклическое выполнение Ping и возвращать nil, если был вызван метод Stop.

func (s *ServiceKeeper) Watch(ctx context.Context) error {
	if !s.checkState(srvStateReady, srvStateRunning) {
		return ErrWrongState
	}
	if err := s.cycleTestServises(ctx); err != nil && err != ErrShutdown {
		return err
	}
	return nil
}

func (s *ServiceKeeper) Stop() {
	if s.checkState(srvStateRunning, srvStateShutdown) {
		close(s.stop)
	}
}

func (s *ServiceKeeper) cycleTestServises(ctx context.Context) error {
	for {
		select {
		case <-s.stop:
			return nil
		case <-time.After(s.PingPeriod):
			if err := s.testServises(ctx); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s *ServiceKeeper) testServises(ctx context.Context) error {
	var ctxPing, cancel = context.WithTimeout(ctx, s.PingTimeout)
	defer cancel()
	var p paralleRun
	for i := range s.Services {
		p.do(ctxPing, s.Services[i].Ping)
	}
	return p.wait()
}

func (s *ServiceKeeper) release() error {
	shCtx, cancel := context.WithTimeout(context.Background(), s.ShutdownTimeout)
	defer cancel()
	var p paralleRun
	for i := range s.Services {
		var service = s.Services[i]
		p.do(shCtx, func(_ context.Context) error {
			return service.Close()
		})
	}
	var errCh = make(chan error)
	go func() {
		defer close(errCh)
		if err := p.wait(); err != nil {
			errCh <- err
		}
	}()
	for {
		select {
		case err, ok := <-errCh:
			if ok {
				return err
			}
			return nil
		case <-shCtx.Done():
			return shCtx.Err()
		}
	}
}

// func (s *ServiceKeeper) release() error {
// 	shCtx, cancel := context.WithTimeout(context.Background(), s.ShutdownTimeout)
// 	defer cancel()
// 	var errCh = make(chan error, len(s.Services))
// 	var wg sync.WaitGroup
// 	wg.Add(len(s.Services))
// 	for i := range s.Services {
// 		go func(service Service) {
// 			defer wg.Done()
// 			if err := service.Close(); err != nil {
// 				errCh <- err
// 			}
// 		}(s.Services[i])
// 	}
// 	go func() {
// 		wg.Wait()
// 		close(errCh)
// 	}()
// 	select {
// 	case err, ok := <-errCh:
// 		if ok {
// 			return err
// 		}
// 		return nil
// 	case <-shCtx.Done():
// 		return shCtx.Err()
// 	}
// }

func (s *ServiceKeeper) Release() error {
	if s.checkState(srvStateShutdown, srvStateOff) {
		return s.release()
	}
	return ErrWrongState
}
