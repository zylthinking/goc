package goc

import (
	"reflect"
	"runtime"
	"sync"
	"unsafe"
)

func getg() unsafe.Pointer
func getg_it() interface{}

func goidImpl() func() int64 {
	var once sync.Once
	var offset int64 = -1
	var fn func() int64
	once.Do(func() {
		if runtime.GOARCH == "386" || runtime.GOARCH == "amd64" {
			it := getg_it()
			if f, ok := reflect.TypeOf(it).FieldByName("goid"); ok {
				offset = int64(f.Offset)
			}
		}

		fn = func() int64 {
			if offset == -1 {
				return -1
			}
			return *(*int64)(unsafe.Pointer(uintptr(getg()) + uintptr(offset)))
		}
	})
	return fn
}

func Goid() int64 {
	fn := goidImpl()
	return fn()
}
