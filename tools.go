package goc

import "reflect"

func ifaceToType(it interface{}, typ reflect.Type) interface{} {
	if it == nil {
		return reflect.New(typ).Elem().Interface()
	}

	defer func() {
		x := recover()
		if x != nil {
			it = reflect.New(typ).Elem().Interface()
		}
	}()

	v := reflect.ValueOf(it)
	kv := v.Convert(typ)
	it = kv.Interface()
	return it
}
