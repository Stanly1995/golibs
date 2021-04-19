package cerr

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

type New string

func (e New) Error() string {
	return string(e)
}

func ErrPerformReq(err error) error {
	return fmt.Errorf("failed to perform request: %w", err)
}

func ErrBadRespCode(response *http.Response) error {
	b, _ := ioutil.ReadAll(response.Body)
	return fmt.Errorf("responded with bad code: %d message: %s", response.StatusCode, string(b))
}

func ErrBadRespBody(response *http.Response, err error) error {
	b, _ := ioutil.ReadAll(response.Body)
	return fmt.Errorf("responded with bad body: %s .Where body: %s", err, string(b))
}

const (
	ErrNotFound = New("Not Found")
)
