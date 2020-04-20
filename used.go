package goc

import (
	"fmt"
	"sync"
	"time"
)

type trace_point struct {
	msg string
	tms int64
	idx int
	end *trace_point
}

type trace struct {
	points []*trace_point
}

var traces sync.Map
func Goid() uintptr

func trace_get() *trace {
	var zt *trace
	gid := Goid()
	it, ok := traces.Load(gid)
	if !ok {
		zt = &trace{}
        traces.Store(gid, zt)
	} else {
		zt = it.(*trace)
	}
	return zt
}

func Mark(msg string) func(...string) int64 {
	zt := trace_get()
	pt1 := &trace_point{
		msg: msg,
		idx: len(zt.points),
		tms: time.Now().UnixNano() / 1000000,
	}
	zt.points = append(zt.points, pt1)

	f := func(msg ...string) int64 {
		pt2 := &trace_point{
			idx: len(zt.points),
			tms: time.Now().UnixNano() / 1000000,
		}

		if len(msg) > 0 {
			pt1.msg += ", " + msg[0]
		}
		zt.points = append(zt.points, pt2)
		pt1.end = pt2
		return pt2.tms - pt1.tms
	}
	return f
}

func (this *trace) explain(start, end int, space string) {
	for i := start; i < end; i++ {
		pt := this.points[i]
		if pt.end == nil {
			fmt.Println(space, pt.msg, pt.tms, "-->")
			continue
		}

		fmt.Println(space, pt.msg, pt.tms, "-->", pt.end.tms, "used", pt.end.tms-pt.tms)
		if pt.end.idx > i+1 {
			this.explain(i+1, pt.end.idx, space+"    ")
		}
		i = pt.end.idx
	}
}

func Show(msg string) {
	zt := trace_get()
	fmt.Println("ztrace\nztrace")
	zt.explain(0, len(zt.points), "ztrace "+msg+": ")
}

func Reset() {
	zt := trace_get()
	zt.points = zt.points[0:0]
}
