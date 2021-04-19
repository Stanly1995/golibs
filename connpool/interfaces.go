package connpool

import (
	"github.com/gorilla/websocket"
	"net/http"
	"time"
)

//go:generate mockery -name IConn -case underscore  -inpkg -testonly
// IConn is a wrapper of websocket connection used
// for messages between FE and BE and between BEs
type IConn interface {
	Send(msg []byte) error
	CloseCb(cb func(connUUID string))
	PingWait(wait int) error
	ReceiveCb(cb func(msg []byte, connUUID string))
	PingMessage(msg string) error
}

//go:generate mockery -name iCloseCallbacksContainer -case underscore  -inpkg -testonly
type iCloseCallbacksContainer interface {
	AddCloseCb(closeCb func(connID string)) error
	CallCloseCbs(connID string)
}

//go:generate mockery -name IConnPool -case underscore -inpkg -testonly
type IConnPool interface {
	Send(msg []byte, connUUID string) error
	Register(conn IConn, connUUID string) error
	ReceiveCb(cb func(msg []byte, connUUID string))
}

//go:generate mockery -name IWs -case underscore -inpkg -testonly
// IWs used for testing purpose
type IWs interface {
	Close() error
	ReadMessage() (messageType int, p []byte, err error)
	CloseHandler() func(code int, text string) error
	WriteMessage(messageType int, data []byte) error
	SetReadDeadline(t time.Time) error
	SetCloseHandler(h func(code int, text string) error)
}

//go:generate mockery -name IDialer -case underscore -inpkg -testonly
// IDialer used for testing purpose
type IDialer interface {
	Dial(urlStr string, requestHeader http.Header) (*websocket.Conn, *http.Response, error)
}

//go:generate mockery -name IUpgrader -case underscore -inpkg -testonly
// IUpgrader used for testing purpose
type IUpgrader interface {
	Upgrade(w http.ResponseWriter, r *http.Request, responseHeader http.Header) (*websocket.Conn, error)
}
