package di

import (
	"context"
	"fmt"
	"reflect"
	"runtime/debug"
	"strings"
)

const getMethodPrefix = "Get"

func Build(ctx context.Context, container any) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("build container: %v", e)

			debug.PrintStack()
		}
	}()

	elem, t := reflect.ValueOf(container), reflect.TypeOf(container)
	for i := 0; i < elem.NumMethod(); i++ {
		if m := t.Method(i); strings.HasPrefix(m.Name, getMethodPrefix) {
			elem.MethodByName(m.Name).Call([]reflect.Value{reflect.ValueOf(ctx)})
		}
	}

	return err
}
