package syncx

import "sync"

func Once(fn func()) func() {
	var once sync.Once

	return func() {
		once.Do(fn)
	}
}
