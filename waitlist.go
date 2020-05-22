package goc

import (
	"sync"
	"sync/atomic"
	"time"
)

type wait_node struct {
	entry   ListHead
	cond    *sync.Cond
	expired int32
}

type WaitList struct {
	head ListHead
	mutx sync.Mutex
	N    int32
}

func WaitListInit(wl *WaitList) {
	InitListHead(&wl.head, &wl)
}

func NeWaitList() *WaitList {
	wl := &WaitList{}
	WaitListInit(wl)
	return wl
}

func WaitOn(wl *WaitList, intp ...int32) (int32, bool) {
	var expire, n int32 = -1, wl.N
	switch len(intp) {
	case 2:
		expire = intp[1]
		fallthrough
	case 1:
		n = intp[0]
	}

	var wn wait_node
	//
	// 以下代码被注释：
	// if (wl.N != n) {
	//     return wl.N, false;
	// }
	// 原因：必须通过 mutex 的内存屏障作用保证 WaitOn 返回后能观察到 Wakeup 之前的更改
	//
	wl.mutx.Lock()
	if wl.N != n {
		goto LABEL
	}

	if expire == 0 {
		wn.expired = 1
		goto LABEL
	}

	if expire > 0 {
		time.AfterFunc(time.Millisecond*time.Duration(expire), func() {
			if atomic.CompareAndSwapInt32(&wn.expired, 0, 1) {
				wl.mutx.Lock()
				ListDel(&wn.entry)
				wl.mutx.Unlock()
				wn.cond.Signal()
			}
		})
	}

	InitListHead(&wn.entry, &wn)
	wn.cond = sync.NewCond(&wl.mutx)
	ListAddTail(&wn.entry, &wl.head)
	wn.cond.Wait()
LABEL:
	n = wl.N
	wl.mutx.Unlock()
	return n, (wn.expired == 1)
}

func Wakeup(wl *WaitList, nr int32) int32 {
	var n int32
	var head ListHead
	InitListHead(&head)
	wl.mutx.Lock()
	if nr == 0 {
		// 不唤醒正在休眠的
		// 但允许即将休眠的直接退出休眠过程
		// 这不是一个可能出现的情况
		// 但 nr == 0 做调用也本不可能出现
		nr = 1
		goto LABEL
	}

	if nr < 0 {
		ListAdd(&head, &wl.head)
		ListDelInit(&wl.head)
	}

	for nr > 0 && !ListEmpty(&wl.head) {
		ent := wl.head.Next
		ListDel(ent)
		ListAdd(ent, &head)
		nr--
	}

LABEL:
	if nr != 0 {
		wl.N++
		n = wl.N
	}
	wl.mutx.Unlock()

	for !ListEmpty(&head) {
		ent := head.Next
		wn := ListEntry(ent).(*wait_node)

		if atomic.CompareAndSwapInt32(&wn.expired, 0, 2) {
			ListDel(&wn.entry)
			wn.cond.Signal()
		}
	}
	return n
}
