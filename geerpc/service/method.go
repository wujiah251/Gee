package service

import (
	"reflect"
	"sync/atomic"
)

type MethodType struct {
	method    reflect.Method //方法本身
	ArgType   reflect.Type   // 第一个参数的类型，arg
	ReplyType reflect.Type   // 第二个参数的类型，reply
	numCalls  uint64         // 后续统计方法调用次数会用到
}

func (m *MethodType) NumCalls() uint64 {
	return atomic.LoadUint64(&m.numCalls)
}

// 传参可以是指针传递或者值传递
func (m *MethodType) NewArgv() reflect.Value {
	var argv reflect.Value
	if m.ArgType.Kind() == reflect.Ptr {
		argv = reflect.New(m.ArgType.Elem())
	} else {
		argv = reflect.New(m.ArgType).Elem()
	}
	return argv
}

// reply可以是
func (m *MethodType) NewReplyv() reflect.Value {
	replyv := reflect.New(m.ReplyType.Elem())
	switch m.ReplyType.Elem().Kind() {
	case reflect.Map:
		replyv.Elem().Set(reflect.MakeMap(m.ReplyType.Elem()))
	case reflect.Slice:
		replyv.Elem().Set(reflect.MakeSlice(m.ReplyType.Elem(), 0, 0))
	}
	return replyv
}
