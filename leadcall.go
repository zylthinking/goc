package goc

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"unsafe"
)

type result struct {
	// mux 及相关被注释掉的代码并不是废弃的
	// 而是看上去 golang atomic 系列函数具有全屏障语义
	// 因此导致 mux 实际上可以被忽略
	//
	// 这里的问题是以下代码
	//   this.value.result, this.value.err = handler()
	// 	     if this.value.result == nil && this.value.err == nil {
	//	     this.value.err = bug
	//   }
	// 和代码
	//   atomic.StorePointer((*unsafe.Pointer)(rawptr), nil)
	//
	// 观察顺序不确定, 因此看上去必须依靠 sync.Mutex 来提供内存屏障
	// 然而阅读 mutex 实现, 其 lock/unlock 本身可退化为 atomic 操作
	// 然而作为一个 mutex, readRequire/writeRelease 是一个基本要求
	// 那么唯一解释就是 atomic 函数本身就提供了全屏障语义
	// 既然如此, 这里也可以直接依赖 atomic 就行
	// 使用 mutex 的保守实现以注释形式提供
	mux    sync.Mutex
	err    error
	result interface{}
}

type LeaderCall struct {
	mux   sync.Mutex
	wl    *WaitList
	value *result
}

type funcPtr = func() (interface{}, error)
var bug = errors.New("bugs in lcHandler")

func (this *LeaderCall) realCall(handler funcPtr) {
	//this.value.mux.Lock()
	this.value.result, this.value.err = handler()
	if this.value.result == nil && this.value.err == nil {
		this.value.err = bug
	}
	//this.value.mux.Unlock()

	rawptr := unsafe.Pointer(&this.value)
	this.mux.Lock()
	atomic.StorePointer((*unsafe.Pointer)(rawptr), nil)
	this.mux.Unlock()
	Wakeup(this.wl, -1)
}

func (this *LeaderCall) EnterCallGate(expire int32, handler funcPtr) (interface{}, error, bool) {
	var val result
	var ptr *result

	this.mux.Lock()
	if this.value == nil {
		this.value = &val
	}
	ptr = this.value
	this.mux.Unlock()

	N := this.wl.N
	var expired bool
	if ptr == &val {
		if expire < 0 {
			this.realCall(handler)
		} else {
			go this.realCall(handler)
		}
	}

	rawptr := unsafe.Pointer(&this.value)
	resultPtr := (*result)(atomic.LoadPointer((*unsafe.Pointer)(rawptr)))
	if ptr == resultPtr {
		_, expired = WaitOn(this.wl, N, expire)
		// WaitOn 中存在内存屏障可保证
		// 至少 this.value = nil 能被观察到
		// 同时顺带保证了 this.value.result/err 被观察到
		// 其他的写不影响下面的判断
		resultPtr = this.value
	} else {
		//ptr.mux.Lock()
		//defer ptr.mux.Unlock()
	}

	if ptr != resultPtr {
		expired = false
	}

	err := ptr.err
	it := ptr.result
	if it == nil {
		fmt.Printf("%p %p\n", ptr, this.value)
		panic(fmt.Sprintf("Bug detected %v", ptr == &val))
	}
	return it, err, expired
}

func NewLeadCall() *LeaderCall {
	call := &LeaderCall{}
	call.wl = NeWaitList()
	return call
}
