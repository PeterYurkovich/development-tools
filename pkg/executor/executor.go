package executor

type UpdateStatus int

const (
	StatusPending UpdateStatus = iota
	StatusInProgress
	StatusComplete
	StatusFailed
)

type ProgressUpdate struct {
	Step    string
	Status  UpdateStatus
	Error   error
	Index   int
	Message string
}

type Executor struct {
	UpdateCh chan ProgressUpdate
}

func NewExecutor() *Executor {
	return &Executor{
		UpdateCh: make(chan ProgressUpdate, 10),
	}
}

func (e *Executor) SendUpdate(index int, status UpdateStatus, step string) {
	e.UpdateCh <- ProgressUpdate{
		Index:  index,
		Status: status,
		Step:   step,
	}
}

func (e *Executor) SendUpdateWithError(index int, status UpdateStatus, step string, err error) {
	e.UpdateCh <- ProgressUpdate{
		Index:  index,
		Status: status,
		Step:   step,
		Error:  err,
	}
}

func (e *Executor) SendLog(index int, message string) {
	e.UpdateCh <- ProgressUpdate{
		Index:   index,
		Status:  StatusInProgress,
		Message: message,
	}
}

func (e *Executor) Close() {
	close(e.UpdateCh)
}
