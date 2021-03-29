package goc

import (
	"strconv"
	"sync"
	"sync/atomic"
	"time"
	_ "unsafe"

	logger "git-biz.qianxin-inc.cn/infra-components/sdk/microservice-framework/go-framework.git/log"
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
var seq int64

func trace_get(gen bool) *trace {
	var zt *trace
	gid := Goid()
	if gid == -1 {
		return nil
	}

	it, ok := traces.Load(gid)
	if !ok {
		if gen {
			zt = &trace{}
			traces.Store(gid, zt)
		}
	} else {
		zt = it.(*trace)
	}
	return zt
}

func trace_put() {
	gid := Goid()
	if gid == -1 {
		return
	}
	traces.Delete(gid)
}

func Mark(msg string, gen ...bool) func(...string) int64 {
	zt := trace_get(len(gen) > 0)
	if zt == nil {
		return func(...string) int64 {
			return -1
		}
	}

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
			logger.Info(space, pt.msg, pt.tms, "-->")
			continue
		}

		logger.Info(space, pt.msg, pt.tms, "-->", pt.end.tms, "used", pt.end.tms-pt.tms)
		if pt.end.idx > i+1 {
			this.explain(i+1, pt.end.idx, space+"    ")
		}
		i = pt.end.idx
	}
}

func Finish(pfx, msg string, print bool) {
	zt := trace_get(false)
	if zt == nil {
		return
	}

	if print {
		strN := strconv.FormatInt(atomic.AddInt64(&seq, 1), 10)
		logger.Info(pfx + "\n" + pfx)
		zt.explain(0, len(zt.points), pfx+" "+msg+"-"+strN+": ")
	}
	zt.points = zt.points[0:0]
	trace_put()
}
