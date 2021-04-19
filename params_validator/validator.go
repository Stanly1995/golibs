package params_validator

import (
	"fmt"
	"github.com/Stanly1995/golibs/cerr"
	"gopkg.in/go-playground/validator.v9"
	"path/filepath"
	"reflect"
	"runtime"
)

const ErrInvalidParam = cerr.New("param is invalid")

func ValidateParamsWithPanic(params ...interface{}) {
	for i := range params {
		if name, err := validateParam(params[i]); err != nil {
			_, fn, line, _ := runtime.Caller(1)
			file := filepath.Base(fn)
			panic(fmt.Sprintf("filename: %s, line: %d, param: %s, error: %v", file, line, name, err))
		}
	}
}

func validateParam(param interface{}) (string, error) {
	if param == nil {
		return "nil", ErrInvalidParam
	}
	paramType := reflect.TypeOf(param)
	paramValue := reflect.ValueOf(param)

	switch paramType.Kind() {
	case reflect.Ptr, reflect.Func:
		if paramValue.IsNil() {
			return paramValue.String(), ErrInvalidParam
		}
	case reflect.Struct:
		return paramValue.String(), validator.New().Struct(param)
	}

	return paramValue.String(), nil
}
