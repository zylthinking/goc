package goc

import (
	"context"
	"sync"
	"unsafe"
)

type execUnit struct {
	ctx    context.Context
	cancel context.CancelFunc
	uptr   []unsafe.Pointer
}

type BgExecutor struct {
	sync.Mutex
	fn    func(context.Context, ...unsafe.Pointer)
	units [2]*execUnit
	mux   *sync.Mutex
}

func NewBgExecutor(fn func(context.Context, ...unsafe.Pointer), mutx ...*sync.Mutex) *BgExecutor {
	var mux *sync.Mutex
	if len(mutx) > 0 {
		mux = mutx[0]
	}

	return &BgExecutor{
		fn:  fn,
		mux: mux,
	}
}

func (this *BgExecutor) Exec(ctx context.Context, uptr ...unsafe.Pointer) context.CancelFunc {
	var unit = &execUnit{
		uptr: uptr,
	}
	unit.ctx, unit.cancel = context.WithCancel(ctx)

	cancel := func() {
		unit.cancel()
		this.Lock()
		defer this.Unlock()

		if this.units[1] == unit {
			this.units[1] = nil
		}
	}

	this.Lock()
	if this.units[0] == nil {
		this.units[0] = unit
		go this.main(unit.ctx, uptr, cancel)
	} else {
		this.units[1] = unit
	}
	this.Unlock()
	return cancel
}

func (this *BgExecutor) main(ctx context.Context, uptr []unsafe.Pointer, cancel context.CancelFunc) {
	if this.mux != nil {
		this.mux.Lock()
		defer this.mux.Unlock()
	}

	for ctx != nil {
		this.fn(ctx, uptr...)
		cancel()

		this.Lock()
		this.units[0], this.units[1] = this.units[1], nil
		if this.units[0] != nil {
			ctx = this.units[0].ctx
			uptr = this.units[0].uptr
			cancel = this.units[0].cancel
		} else {
			ctx = nil
		}
		this.Unlock()
	}
}
