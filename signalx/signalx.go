package signalx

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/byteflowing/go-common/service"
)

type SignalHandler interface {
	Start()
	Stop()
}

type SignalListener struct {
	handlers []SignalHandler
	funcs    []func()

	waitDuration time.Duration
}

func NewSignalListener(wait time.Duration) *SignalListener {
	return &SignalListener{
		waitDuration: wait,
	}
}

func (s *SignalListener) Register(handler SignalHandler) {
	s.handlers = append(s.handlers, handler)
}

func (s *SignalListener) RegisterFunc(stop func()) {
	s.funcs = append(s.funcs, stop)
}

func (s *SignalListener) Listen() {
	for _, handler := range s.handlers {
		h := handler
		go func() {
			h.Start()
		}()
	}
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer func() {
		signal.Stop(sigCh)
		close(sigCh)
	}()

	sig := <-sigCh
	log.Printf("[signalx] received signal: %v", sig)
	stopGroup := service.NewRoutineGroup()
	for _, handler := range s.handlers {
		h := handler
		stopGroup.Run(func() {
			h.Stop()
		})
	}
	for _, stopFunc := range s.funcs {
		s := stopFunc
		stopGroup.Run(func() {
			s()
		})
	}
	done := make(chan struct{})
	go func() {
		stopGroup.Wait()
		close(done)
	}()
	select {
	case <-done:
		log.Printf("[signalx] stopped gracefully")
	case sig2 := <-sigCh:
		log.Printf("[signalx] received second signal: %v, force kill", sig2)
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGKILL)
	case <-time.After(s.waitDuration):
		log.Printf("[signalx] timeout after %v secods, force kill", s.waitDuration)
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGKILL)
	}
}
