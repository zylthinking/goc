package goc

import (
	"runtime"
	"sync"
	"sync/atomic"
	"unsafe"
)

type cond struct {
	uptr unsafe.Pointer
	mux  sync.Mutex
	cond *sync.Cond
	nr   [2]int64
}

func newCond() *cond {
	cond := &cond{}
	cond.cond = sync.NewCond(cond)
	return cond
}

func (this *cond) Lock() {
	this.mux.Lock()
}

func (this *cond) Unlock() {
	this.nr[1]--
	this.mux.Unlock()
}

func (this *cond) wait(nr int64) {
	this.Lock()
	this.nr[0] += nr
	this.nr[1]++
	this.cond.Wait()
	this.Unlock()
}

func (this *cond) wake(uptr unsafe.Pointer) {
	this.uptr = uptr
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
		cond.wait(nr)
		uptr = unsafe.Pointer(&cond.uptr)
	} else {
		this.mux.Lock()
		this.holder = (*cond)(atomic.SwapPointer(ptr, unsafe.Pointer(newCond())))
		nr = atomic.SwapInt64(&this.nr, 0) - nr
		for this.holder.nr[0] != nr || this.holder.nr[1] != 0 {
			runtime.Gosched()
		}
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
