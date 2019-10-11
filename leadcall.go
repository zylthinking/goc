package goc

import (
    "errors"
)

type LeaderCall struct {
    pctx PContext;
    wl WaitList;
    err error;
    result interface{};
}
type lcHandler func() (interface{}, error);

type lcNode struct {
    lkn LkfNode;
    handler func(*LeaderCall);
}
var bug = errors.New("bugs in lcHandler");

func (this *LeaderCall) handlerWrapper(handler lcHandler) func(*LeaderCall) {
    return func(lc *LeaderCall) {
        if (lc.err == nil && lc.result == nil) {
            lc.result, lc.err = handler();
            if (lc.result == nil && lc.err == nil) {
                lc.err = bug;
            }
            Wakeup(&this.wl, -1);
        }
    }
}

func (this *LeaderCall) realCall() {
    var lkn *LkfNode;
LABEL:
    lkn = LkfNodeGet(&this.pctx.List);

    var current *LkfNode;
    for (current != lkn) {
        current = LkfNodeNext(lkn);
        if (current != nil) {
            lcn := current.Container.(*lcNode);
            lcn.handler(this);
        }
    }

    if (!ProcLeave(&this.pctx)) {
        goto LABEL;
    }
}

func (this *LeaderCall) EnterCallGate(expire int32, handler lcHandler) (interface{}, error, bool) {
    lcn := &lcNode{};
    LknInit(&lcn.lkn, lcn);
    lcn.handler = this.handlerWrapper(handler);

    n := this.wl.N;
    b := ProcEnter(&this.pctx, &lcn.lkn);
    if (b) {
        go this.realCall();
    }

    if (this.result == nil && this.err == nil) {
        _, b = WaitOn(&this.wl, n, expire);
    }

    if (this.result != nil || this.err != nil) {
        b = false;
    }
    return this.result, this.err, b;
}

func NewLeadCall() *LeaderCall {
    call := &LeaderCall{};
    WaitListInit(&call.wl);
    ProcInit(&call.pctx);
    return call;
}
