package goc

import (
    "sync/atomic"
    "unsafe"
    "runtime"
)

type LkfNode struct {
    next unsafe.Pointer;
    Container interface{};
}

func LknInit(node *LkfNode, container interface{}) {
    node.Container = container;
}

type LkfList struct  {
    root LkfNode ;
    tail unsafe.Pointer;
}

func init() {
    var node LkfNode;
    lkf_node_offset = unsafe.Offsetof(node.next);
}
var lkf_node_offset uintptr = 0;

func LkfInit(list *LkfList, addr ...interface{}) {
    list.root.next = nil;
    if (len(addr) != 0) {
        LknInit(&list.root, addr[0]);
    }
    list.tail = unsafe.Pointer(&list.root.next);
}

func LkfNodePut(list *LkfList, node *LkfNode) {
    node.next = nil;
    ptr := atomic.SwapPointer(&list.tail, unsafe.Pointer(&node.next));
    *(*unsafe.Pointer)(ptr) = unsafe.Pointer(node);
}

func LkfNodeGet(list *LkfList) *LkfNode {
    first := atomic.SwapPointer(&list.root.next, nil);
    ptr := (*LkfNode)(first);
    if (ptr == nil) {
        return nil;
    }

    last := atomic.SwapPointer(&list.tail, unsafe.Pointer(&list.root.next));
    *(*unsafe.Pointer)(last) = first;

    ptr = (*LkfNode)(unsafe.Pointer(uintptr(last) - lkf_node_offset));
    runtime.KeepAlive(last);
    return ptr;
}

func LkfNodeNext(node *LkfNode) *LkfNode {
    ptr := (*LkfNode)(node.next);
    if (ptr == nil || ptr.next == nil) {
        return nil;
    }

    if (ptr != node) {
        node.next = ptr.next;
    }
    ptr.next = nil;
    return ptr;
}

type PContext struct {
    List LkfList;
    stat int32;
}

func ProcInit(ctx *PContext) {
    LkfInit(&ctx.List, ctx);
}

func ProcEnter(ctx *PContext, node *LkfNode) bool {
    LkfNodePut(&ctx.List, node);
    return atomic.CompareAndSwapInt32(&ctx.stat, 0, 1);
}

func ProcLeave(ctx *PContext) bool {
    atomic.StoreInt32(&ctx.stat, 0);
    tail := (*unsafe.Pointer)(atomic.LoadPointer(&ctx.List.tail));
    if (tail == &ctx.List.root.next) {
        return true;
    }
    return !atomic.CompareAndSwapInt32(&ctx.stat, 0, 1);
}
