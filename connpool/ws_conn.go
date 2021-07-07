package connpool

import (
	"github.com/Stanly1995/golibs/cerr"
	"github.com/labstack/gommon/log"
	"reflect"
	"time"
)

const (
	// Time allowed to read the next ping message from the peer.
	pingWait = 20 * time.Second
	// pingMessage is body of incoming ping message
	pingMessage       = "."
	maxPingMessageLen = 4
	maxPingWait       = 20 * time.Second
)

// WsConn is a new websocket client/server wrapper
type WsConn struct {
	conn        IWs
	uuid        string
	receiveCb   func(msg []byte, connUUID string)
	closeCb     func(connUUID string)
	pingMessage string
	pingWait    time.Duration
}

var timeNow = func() time.Time {
	return time.Now()
}

// InitAndRunWsConn server WsConn
// and starts async loop which a message
// to a receiveCb function
func InitAndRunWsConn(ws IWs, connUUID string) (*WsConn, error) {
	log.Debugf("ws client %s has connected", connUUID)
	if reflect.ValueOf(ws).IsNil() {
		return nil, cerr.ErrFuncArg{}.Invalidate("ws")
	}
	if connUUID == "" {
		return nil, cerr.ErrFuncArg{}.Invalidate("connUUID")
	}
	wsc := &WsConn{
		uuid:        connUUID,
		conn:        ws,
		receiveCb:   func(msg []byte, connUUID string) {},
		pingMessage: pingMessage,
		pingWait:    maxPingWait,
	}
	wsc.runReceiveLoop()
	return wsc, nil
}

func (wsc *WsConn) runReceiveLoop() {
	err := wsc.conn.SetReadDeadline(timeNow().Add(maxPingWait))
	if err != nil {
		log.Error(err)
		wsc.close()
		return
	}
	go func() {
		for {
			_, msg, err := wsc.conn.ReadMessage()
			if err != nil {
				log.Errorf("Failed to read from ws: %v", err)
				wsc.close()
				return
			}
			if wsc.pingMessage != "" && string(msg) == wsc.pingMessage {
				err = wsc.conn.SetReadDeadline(time.Now().Add(wsc.pingWait))
				if err != nil {
					log.Error(err)
				}
				continue
			}
			go wsc.receiveCb(msg, wsc.uuid)
		}
	}()
}

// CloseCb sets a callback which will be called
// when connection from client side was closed.
func (wsc *WsConn) CloseCb(cb func(connUUID string)) {
	wsc.closeCb = cb
	standartHandler := wsc.conn.CloseHandler()
	wsc.conn.SetCloseHandler(func(code int, text string) error {
		cb(wsc.uuid)
		return standartHandler(code, text)
	})
}

func (wsc *WsConn) close() {
	err := wsc.conn.Close()
	if err != nil {
		log.Error(err)
	}
	wsc.closeCb(wsc.uuid)
}

func (wsc *WsConn) Close() {
	wsc.close()
}

// ReceiveCb sets a callback which will be called
// when message from client side was received.
func (wsc *WsConn) ReceiveCb(cb func(msg []byte, connUUID string)) {
	wsc.receiveCb = cb
}

// Send sends message to client side
func (wsc *WsConn) Send(msg []byte) error {
	err := wsc.conn.WriteMessage(1, msg)
	if err != nil {
		wsc.close()
	}
	return err
}

// PingMessage is setter for ping message.
// Will return error when msg is empty string.
// There is default value in InitAndRunWsConn func
func (wsc *WsConn) PingMessage(msg string) error {
	if msg == "" || len(msg) > maxPingMessageLen {
		wsc.close()
		return ErrInvalidPingMsg
	}
	wsc.pingMessage = msg
	return nil
}

// PingWait is setter for ping timeout.
// Will return error when msg is empty string.
// There is default value in InitAndRunWsConn func
func (wsc *WsConn) PingWait(wait int) error {
	newPingWait := time.Second * time.Duration(wait)
	if wait < 1 || newPingWait.Seconds() > maxPingWait.Seconds() {
		wsc.close()
		return ErrInvalidPingWait
	}
	wsc.pingWait = newPingWait
	return nil
}
