package goc

import "reflect"

func Iface2Type(it interface{}, typ reflect.Type) (get interface{}) {
    if it == nil {
        return reflect.New(typ).Elem().Interface()
    }   

    defer func() {
        x := recover()
        if x != nil {
            get = reflect.New(typ).Elem().Interface()
        }
    }() 

    v := reflect.ValueOf(it)
    kv := v.Convert(typ)
    return kv.Interface()
}
