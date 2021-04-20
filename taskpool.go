package goc

type TaskPool struct {
	in    chan *func()
	busy  chan *func()
	keep  int32
	idle  int32
	total int32
}

func NewTaskPool(keep int32) *TaskPool {
	if keep < 0 {
		keep = 0
	}

	pool := &TaskPool{
		in:   make(chan *func()),
		busy: make(chan *func(), 8),
		keep: keep,
	}
	go pool.main()
	return pool
}

func (this *TaskPool) main() {
	for {
		fptr := <-this.busy
		atomic.AddInt32(&this.total, 1)
		go this.execute(fptr)
	}
}

func (this *TaskPool) execute(fptr *func()) {
	for {
		(*fptr)()
		n := atomic.AddInt32(&this.idle, 1)
		if n > this.keep {
			atomic.AddInt32(&this.idle, -1)
			atomic.AddInt32(&this.total, -1)
			return
		}
		fptr = <-this.in
		atomic.AddInt32(&this.idle, -1)
	}
}

func (this *TaskPool) SubmitTask(fptr *func()) {
	select {
	case this.in <- fptr:
		return
	default:
		if this.total > 4096 {
			this.in <- fptr
			return
		}
	}

	select {
	case this.in <- fptr:
	case this.busy <- fptr:
	}
}
