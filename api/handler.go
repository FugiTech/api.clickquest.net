package api

import (
	"log"
	"reflect"
)

var handlers = map[string]*handler{}

type handler struct {
	Params reflect.Type
	Func   func(*Conn, interface{}) (interface{}, error)
	after  []func(*Conn)
}

func RegisterHandler(name string, params interface{}, fn func(*Conn, interface{}) (interface{}, error)) *handler {
	h := &handler{
		Func: fn,
	}

	if params != nil {
		t := reflect.TypeOf(params)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		if t.Kind() != reflect.Struct {
			log.Panicf("Invalid type for handler %q: %v", name, t)
		}
		h.Params = t
	}

	handlers[name] = h
	return h
}

func (h *handler) After(fn func(*Conn)) {
	h.after = append(h.after, fn)
}

func (h *handler) RunAfter(c *Conn) {
	for _, fn := range h.after {
		fn(c)
	}
}
