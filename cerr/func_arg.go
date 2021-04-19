package cerr

import (
	"fmt"
	"runtime"
	"strings"
)

const invalidArgStr = "%s has invalid %s argument"

// ErrFuncArg represents an error of a function argument validation
type ErrFuncArg struct {
	FuncName string
	Arg      string
}

// NewErrFuncArgMock creates new ErrFuncArg used in a testing
func NewErrFuncArgMock(argName, funcName string) ErrFuncArg {
	return ErrFuncArg{FuncName: funcName}.Invalidate(argName)
}

// Invalidate fires an ErrFuncArg
// which consists of caller's name and an invalid arg name
func (e ErrFuncArg) Invalidate(argName string) ErrFuncArg {
	if e.FuncName == "" {
		pc := make([]uintptr, 1)
		runtime.Callers(2, pc)
		if pc[0] != 0 {
			frame, _ := runtime.CallersFrames(pc).Next()
			s := strings.Split(frame.Function, ".")
			e.FuncName = s[len(s)-1]
		}
	}
	e.Arg = argName
	return e
}

// Error returns error's string
func (e ErrFuncArg) Error() string {
	return fmt.Sprintf(invalidArgStr, e.FuncName, e.Arg)
}
