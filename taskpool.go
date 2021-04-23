package goc

import (
	"sync/atomic"
	"time"
)

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
		busy: make(chan *func(), 64),
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
		fptr = nil
		
		n := atomic.AddInt32(&this.idle, 1)
		if n > this.keep {
			atomic.AddInt32(&this.idle, -1)
			atomic.AddInt32(&this.total, -1)
			return
		}

		select {
		case fptr = <-this.in:
		default:
			timer := time.NewTimer(time.Second * 30)
			select {
			case fptr = <-this.in:
			case <-timer.C:
			}
			timer.Stop()
		}

		atomic.AddInt32(&this.idle, -1)
		if fptr == nil {
			atomic.AddInt32(&this.total, -1)
			return
		}
	}
}

func (this *TaskPool) SubmitTask(fptr *func()) {
	select {
	case this.in <- fptr:
		return
	default:
		if this.total > 10240 {
			this.in <- fptr
			return
		}
	}

	select {
	case this.in <- fptr:
	case this.busy <- fptr:
	}
}
