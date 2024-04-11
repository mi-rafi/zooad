package internal

import (
	"context"
	"fmt"
	"sync"
)

type appError string

const (
	ErrWrongState  appError = "wrong application state"
	ErrMainOmitted appError = "main function is omitted"
	ErrShutdown    appError = "application is in shutdown state"
	ErrTermTimeout appError = "termination timeout"
)

func (e appError) Error() string {
	return string(e)
}

type arrError []error

func (e arrError) Error() string {
	if len(e) == 0 {
		return "smt went wrong"
	}
	var s = "the following errors occurred:"
	for i := range e {
		s += "\n" + e[i].Error()
	}
	return s
}

type paralleRun struct {
	mux sync.Mutex
	wg  sync.WaitGroup
	err arrError
}

func (p *paralleRun) do(ctx context.Context, f func(context.Context) error) {
	p.wg.Add(1)
	go func() {
		defer func() {
			r := recover()
			if r != nil {
				p.mux.Lock()
				p.err = append(p.err, fmt.Errorf("unhandled error: %v", r))
				p.mux.Unlock()
			}
			p.wg.Done()
		}()
		if err := f(ctx); err != nil {
			p.mux.Lock()
			p.err = append(p.err, fmt.Errorf("%w", err))
			p.mux.Unlock()
		}
	}()
}

func (p *paralleRun) wait() error {
	p.wg.Wait()
	if len(p.err) > 0 {
		return p.err
	}
	return nil
}
