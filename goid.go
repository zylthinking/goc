package goc

import (
	"reflect"
	"unsafe"
)

//go:linkname __a runtime.convT2E
func __a()
func getg() unsafe.Pointer
func getg_it() interface{}

var g_goid_offset uintptr = func() uintptr {
	it := getg_it()
	if f, ok := reflect.TypeOf(it).FieldByName("goid"); ok {
		return f.Offset
	}
	panic("can not find g.goid field")
}()

func Goid() int64 {
	g := getg()
	p := (*int64)(unsafe.Pointer(uintptr(g) + g_goid_offset))
	return *p
}
