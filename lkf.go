package sysapi

import (
    "sync/atomic"
    "unsafe"
    "runtime"
)

type LkfNode struct {
    next *LkfNode;
    Container interface{};
}

type LkfList struct  {
    root LkfNode ;
    tail **LkfNode;
}

func init() {
    var node LkfNode;
    lkf_node_offset = unsafe.Offsetof(node.next);
}
var lkf_node_offset uintptr = 0;

func LkfInit(list *LkfList, addr ...interface{}) {
    list.root.next = nil;
    if (len(addr) != 0) {
        list.root.Container = addr[0];
    }
    list.tail = &list.root.next;
}

func LkfNodePut(list *LkfList, node *LkfNode) {
    node.next = nil;
    tail := unsafe.Pointer(list.tail);
    ptr := atomic.SwapPointer(&tail, unsafe.Pointer(&node.next));
    *(**LkfNode)(ptr) = node;
}

func LkfNodeGet(list *LkfList) *LkfNode {
    next := unsafe.Pointer(list.root.next);
    ptr := (*LkfNode)(atomic.SwapPointer(&next, nil));
    if (ptr == nil) {
        return nil;
    }

    tail := unsafe.Pointer(list.tail);
    last := atomic.SwapPointer(&tail, unsafe.Pointer(&list.root.next));
    *(**LkfNode)(last) = ptr;

    ptr = (*LkfNode)(unsafe.Pointer(uintptr(last) - lkf_node_offset));
    runtime.KeepAlive(last);
    return ptr;
}

func LkfNodeNext(node *LkfNode) *LkfNode {
    ptr := node.next;
    if (ptr == nil || ptr.next == nil) {
        return nil;
    }

    if (ptr == node) {
        return ptr;
    }
    node.next = ptr.next;
    ptr.next = nil;
    return ptr;
}

type PContext struct {
    list LkfList;
    stat int32;
}

func ProcInit(ctx *PContext) {
    LkfInit(&ctx.list, ctx);
}

func ProcEnter(ctx *PContext, node *LkfNode) bool {
    LkfNodePut(&ctx.list, node);
    return atomic.CompareAndSwapInt32(&ctx.stat, 0, 1);
}

func ProcLeave(ctx *PContext) bool {
    atomic.StoreInt32(&ctx.stat, 0);
    if (ctx.list.tail == &ctx.list.root.next) {
        return true;
    }
    return !atomic.CompareAndSwapInt32(&ctx.stat, 0, 1);
}
