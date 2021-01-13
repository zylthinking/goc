package goc

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

type cond struct {
	uptr unsafe.Pointer
	mux  sync.Mutex
	cond *sync.Cond
}

func newCond() *cond {
	cond := &cond{}
	cond.cond = sync.NewCond(&cond.mux)
	return cond
}

func (this *cond) wait() {
	this.mux.Lock()
	for this.uptr == nil {
		this.cond.Wait()
	}
	this.mux.Unlock()
}

func (this *cond) wake(uptr unsafe.Pointer) {
	this.mux.Lock()
	this.uptr = uptr
	this.mux.Unlock()
	this.cond.Broadcast()
}

type LeadLock struct {
	nr     int64
	cond   *cond
	holder *cond
	mux    sync.Mutex
}

func NewLeadLock() *LeadLock {
	leadlock := &LeadLock{}
	leadlock.cond = newCond()
	return leadlock
}

func (this *LeadLock) Lock(nr int64) unsafe.Pointer {
	var uptr unsafe.Pointer
	n := atomic.AddInt64(&this.nr, nr)
	ptr := (*unsafe.Pointer)(unsafe.Pointer(&this.cond))
	if n > 1 {
		cond := (*cond)(atomic.LoadPointer(ptr))
		cond.wait()
		uptr = unsafe.Pointer(&cond.uptr)
	} else {
		this.mux.Lock()
		this.holder = (*cond)(atomic.SwapPointer(ptr, unsafe.Pointer(newCond())))
	}
	return uptr
}

func (this *LeadLock) Unlock(uptr unsafe.Pointer) {
	holder := this.holder
	if uptr == nil || holder == nil {
		return
	}

	this.holder = nil
	holder.wake(uptr)
	this.mux.Unlock()
}
