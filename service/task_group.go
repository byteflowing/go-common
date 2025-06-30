package service

type TaskRunner struct {
	limits chan struct{}
}

func NewTaskRunner(limits int) *TaskRunner {
	return &TaskRunner{limits: make(chan struct{}, limits)}
}

func (t *TaskRunner) Schedule(task func()) {
	t.limits <- struct{}{}

	go func() {
		defer func() {
			<-t.limits
		}()

		task()
	}()
}
