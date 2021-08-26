// +build 386, amd64

package goc

import (
	"reflect"
	"runtime"
	"strings"
	"sync"
	"unsafe"
)

//go:linkname __a runtime.convT2E
func __a()
func getg() unsafe.Pointer
func getg_it() interface{}

var newabi = func() bool {
	var newabi = false
	v := runtime.Version()[2:]
	vs := strings.Split(v, ".")
	if vs[0] > "1" || vs[1] > "16" {
		newabi = true
	}
	return newabi
}()

func goid() int64 {
	var once sync.Once
	var offset int64 = -1
	var fn func() int64

	once.Do(func() {
		it := getg_it()
		if f, ok := reflect.TypeOf(it).FieldByName("goid"); ok {
			offset = int64(f.Offset)
		}

		fn = func() int64 {
			if offset == -1 {
				return -1
			}
			return *(*int64)(unsafe.Pointer(uintptr(getg()) + uintptr(offset)))
		}
	})
	return fn()
}
