package sysapi

import (
    "sync"
    "time"
)

type wait_node struct {
    entry ListHead;
    mutx sync.Mutex;
    cond *sync.Cond;
    expired int32;
}

func (wn *wait_node) wake(expired bool) {
    wn.mutx.Lock();
    if (expired && wn.expired != 0 || !expired && wn.expired == 1) {
        wn.mutx.Unlock();
        return;
    }

    if (wn.expired == 0) {
        wn.expired = 2;
        if (expired) {
            wn.expired = 1;
        }
    }
    ListDel(&wn.entry);
    wn.mutx.Unlock();
    wn.cond.Signal();
}

type WaitList struct {
    head ListHead;
    mutx sync.Mutex;
    N int32;
}

func WaitListInit(wl *WaitList) {
    InitListHead(&wl.head, &wl);
}

func NeWaitList() *WaitList {
    wl := &WaitList{};
    WaitListInit(wl);
    return wl;
}

func WaitOn(wl *WaitList, intp ...int32) (int32, bool) {
    var expire, n int32 = -1, wl.N;
    nr := len(intp);
    if (nr > 0) {
        n = intp[0];
        if (nr > 1) {
            expire = intp[1];
        }
    }

    var wn wait_node;
    if (wl.N != n) {
        return wl.N, false;
    }

    wl.mutx.Lock();
    if (wl.N != n) {
        goto LABEL;
    }

    if (expire == 0) {
        wn.expired = 1;
        goto LABEL;
    }

    if (expire > 0) {
        time.AfterFunc(time.Millisecond * time.Duration(expire), func() {
            if (wn.expired != 0) {
                return;
            }

            wl.mutx.Lock();
            wn.wake(true);
            wl.mutx.Unlock();
        });
    }

    InitListHead(&wn.entry, &wn);
    wn.cond = sync.NewCond(&wl.mutx);
    ListAddTail(&wn.entry, &wl.head);
    wn.cond.Wait();

LABEL:
    n = wl.N;
    wl.mutx.Unlock();
    return n, (wn.expired == 1);
}

func Wakeup(wl *WaitList, nr int32) {
    var head ListHead;
    InitListHead(&head);
    wl.mutx.Lock();
    if (nr == 0) {
        // 不唤醒正在休眠的
        // 但允许即将休眠的直接退出休眠过程
        // 这不是一个可能出现的情况
        // 但 nr == 0 做调用也本不可能出现
        wl.N++;
        wl.mutx.Unlock();
        return;
    }

    if (nr < 0) {
        ListAdd(&head, &wl.head);
        ListDelInit(&wl.head);
    }

    for (nr > 0 && !ListEmpty(&wl.head)) {
        ent := wl.head.Next;
        ListDel(ent);
        ListAdd(ent, &head);
        wn := ListEntry(ent).(*wait_node);
        wn.expired = 2;
        nr--;
    }

    if (ListEmpty(&head) || nr > 0) {
        wl.N++;
    }
    wl.mutx.Unlock();

    for (!ListEmpty(&head)) {
        ent := head.Next;
        wn := ListEntry(ent).(*wait_node);
        wn.wake(false);
    }
}