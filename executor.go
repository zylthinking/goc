package subscriber

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
}

func NewBgExecutor(fn func(context.Context, ...unsafe.Pointer)) *BgExecutor {
	return &BgExecutor{
		fn: fn,
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
		go this.main(unit.ctx, uptr)
	} else {
		this.units[1] = unit
	}
	this.Unlock()
	return cancel
}

func (this *BgExecutor) main(ctx context.Context, uptr []unsafe.Pointer) {
	for ctx != nil {
		this.fn(ctx, uptr...)

		this.Lock()
		this.units[0], this.units[1] = this.units[1], nil
		if this.units[0] != nil {
			ctx = this.units[0].ctx
			uptr = this.units[0].uptr
		} else {
			ctx = nil
		}
		this.Unlock()
	}
}
