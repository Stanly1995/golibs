package cerr

import "strings"

const invalidParamMsg = "Invalid param"

type ErrInvalidParam struct {
	params []string
}

func (eip ErrInvalidParam) AddParam(paramName string) ErrInvalidParam {
	if paramName != "" {
		eip.params = append(eip.params, paramName)
	}
	return eip
}

func (eip ErrInvalidParam) Error() string {
	var errStr strings.Builder
	errStr.WriteString(invalidParamMsg)
	if len(eip.params) > 1 {
		errStr.WriteString("s: ")
	} else {
		errStr.WriteString(": ")
	}
	for i := range eip.params {
		errStr.WriteString(eip.params[i])
		if i != (len(eip.params) - 1) {
			errStr.WriteString(", ")
		}
	}
	return errStr.String()
}
