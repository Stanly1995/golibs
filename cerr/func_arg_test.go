package cerr

import (
	"testing"
)

func mockFunc(testArg string) error {
	if testArg == "" {
		return ErrFuncArg{}.Invalidate("testArg")
	}
	return nil
}

func TestErrFuncArg_StaticFunc(t *testing.T) {
	gotErr := mockFunc("")
	wantErrStr := "mockFunc has invalid testArg argument"

	if wantErrStr != gotErr.Error() {
		t.Error("result doesn't equals")
	}
}
