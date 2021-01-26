package goc

import (
	"context"
	"sync"
)

type execUnit struct {
	ctx    context.Context
	cancel context.CancelFunc
}

type BgExecutor struct {
	sync.Mutex
	fn    func(context.Context)
	units [2]*execUnit
}

func (this *BgExecutor) Exec(ctx context.Context) context.CancelFunc {
	var unit = &execUnit{}
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
		go this.fn(unit.ctx)
	} else {
		this.units[1] = unit
	}
	this.Unlock()
	return cancel
}
