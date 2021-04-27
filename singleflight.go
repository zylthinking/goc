package singleflight

import (
	"runtime"
	"sync"
	"unsafe"
)

type SingleFlight struct {
	sync.Map
}

func (this *SingleFlight) leadLock(key string) *LeadLock {
	iface, ok := this.Load(key)
	if !ok {
		iface, _ = this.LoadOrStore(key, lock.NewLeadLock())
	}
	return iface.(*LeadLock)
}

type result struct {
	v   interface{}
	err error
}

func (this *SingleFlight) Call(key string, fn func() (interface{}, error)) (interface{}, error) {
	var res *result
	defer func() {
		if e, ok := res.err.(*panicError); ok {
			panic(e)
		} else if res.err == errGoexit {
			runtime.Goexit()
		}
	}()

	ll := this.leadLock(key)
	ptr := ll.Lock()
	if ptr != nil {
		res = (*result)(ptr)
		return res.v, res.err
	}

	res = &result{}
	func() {
		var getBack bool
		defer func() {
			if !getBack {
				if r := recover(); r != nil {
					res.err = newPanicError(r)
				} else {
					res.err = errGoexit
				}
			}
			ll.Unlock(unsafe.Pointer(res))
		}()
		res.v, res.err = fn()
		getBack = true
	}()

	return res.v, res.err
}
