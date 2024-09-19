package helper

import (
	"context"
	"reflect"
)

type Service interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

func GetServiceName(s Service) string {
	if o, ok := s.(interface{ Name() string }); ok {
		return o.Name()
	} else {
		t := reflect.TypeOf(s)
		for t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		return t.PkgPath()
	}
}
