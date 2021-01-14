package goc

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"unsafe"
)

type seq_node struct {
	LkfNode
	mux sync.Mutex
}

type SeqLock struct {
	nr     int64
	wl     LkfList
	head   *LkfNode
	holder *seq_node
	mux    sync.Mutex
}

func NewSeqLock() *SeqLock {
	seqlock := &SeqLock{}
	LkfInit(&seqlock.wl)
	return seqlock
}

func (this *SeqLock) Lock() {
	var node seq_node
	LkfNodePut(&this.wl, &node.LkfNode)	
	n := atomic.AddInt64(&this.nr, 1)
	if n > 1 {
		node.mux.Lock()
		node.mux.Lock()
		if this.holder != &node {
			fmt.Println("Bugs in SeqLock")
			os.Exit(1)
		}
	} else {
		this.mux.Lock()
		this.head = LkfNodeGet(&this.wl)
		this.holder = this.next()
	}
}

func (this *SeqLock) next() *seq_node {
	f1 := Mark("seq next")
	defer f1("done")
LABEL:
	lkfn := LkfNodeNext(this.head)
	if lkfn == nil {
		runtime.Gosched()
		goto LABEL
	}
	return (*seq_node)(unsafe.Pointer(lkfn))
}

func (this *SeqLock) Unlock() {
	n := atomic.AddInt64(&this.nr, -1)
	if n == 0 {
		this.mux.Unlock()
		return
	}

	// 本批次最后一个 unlock
	// 但 nr > 0 说明有其他的 Lock 已经进入 wl
	// 获取一个新的批次
	// 可以断言批次一定不为空
	if this.head == &this.holder.LkfNode {
		this.head = LkfNodeGet(&this.wl)
	}
	this.holder = this.next()
	this.holder.mux.Unlock()
}
