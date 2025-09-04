package signalx

import (
	"log"
	"os"
	"os/signal"
	"sync"
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
	mux      sync.Mutex
}

func NewSignalListener() *SignalListener {
	return &SignalListener{}
}

func (s *SignalListener) Register(handler SignalHandler) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.handlers = append(s.handlers, handler)
}

func (s *SignalListener) RegisterFunc(stop func()) {
	s.mux.Lock()
	defer s.mux.Unlock()
	s.funcs = append(s.funcs, stop)
}

func (s *SignalListener) Listen() {
	startGroup := service.NewRoutineGroup()
	for _, handler := range s.handlers {
		h := handler
		startGroup.Run(func() {
			h.Start()
		})
	}
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	log.Println("signalx: received signal: ", sig)
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
		startGroup.Wait()
		close(done)
	}()
	select {
	case <-done:
		stopGroup.Wait()
		signal.Stop(sigCh)
		close(sigCh)
		log.Println("signalx: stopped")
	case sig2 := <-sigCh:
		log.Printf("signalx: received second signal: %v, force kill", sig2)
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGKILL)
	case <-time.After(30 * time.Second):
		log.Printf("signalx: timeout after 30 seconds, force kill")
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGKILL)
	}
}
