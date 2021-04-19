package connpool

import (
	"github.com/Stanly1995/golibs/cerr"
	"sync"
)

const (
	// ErrWrongConnID is error, which is returned when conn id isn't existed in conn pool
	ErrWrongConnID = cerr.New("conn id not found in ConnPool")

	// ErrInvalidCloseCb is error, which is returned when input callback is nil
	ErrInvalidCloseCb = cerr.New("close cb is invalid")

	// ErrInvalidPingMsg is error, which is returned when input ping message is empty
	ErrInvalidPingMsg = cerr.New("invalid ping message")

	// ErrInvalidPingWait is error, which is returned when input ping timeout is 0
	ErrInvalidPingWait = cerr.New("invalid ping wait")
)

// callbacksContainer is needed for storing and calling callbacks when wsConn closes
// it implements iCallbacksContainer
type callbacksContainer struct {
	callbacks []func(connID string)
	mu        sync.Mutex
}

// AddCloseCb adds func(connID string) to callbacks slice
func (cbc *callbacksContainer) AddCloseCb(closeCb func(connID string)) error {
	if closeCb == nil {
		return ErrInvalidCloseCb
	}
	cbc.mu.Lock()
	cbc.callbacks = append(cbc.callbacks, closeCb)
	cbc.mu.Unlock()
	return nil
}

// CallCloseCbs calls all callbacks in callbacks slice
func (cbc *callbacksContainer) CallCloseCbs(connID string) {
	cbc.mu.Lock()
	for _, closeCb := range cbc.callbacks {
		closeCb(connID)
	}
	cbc.mu.Unlock()
}

// ConnPool represents the connpool that holds
// every connection established by client/manager side.
type ConnPool struct {
	iCloseCallbacksContainer
	pool      map[string]IConn
	receiveCb func(msg []byte, connID string)
	mu        sync.Mutex
}

// NewConnPool initializes a new connection connpool and BE-connection connpool
func NewConnPool() *ConnPool {
	cp := &ConnPool{
		pool:      map[string]IConn{},
		receiveCb: func(msg []byte, connID string) {},
	}

	cp.iCloseCallbacksContainer = &callbacksContainer{
		callbacks: []func(connID string){
			cp.unregister,
		},
	}
	return cp
}

// Register adds User to connpool and link close
// and receive events to appropriate callbacks
func (ccp *ConnPool) Register(conn IConn, connID string) error {
	if conn == nil {
		return cerr.ErrFuncArg{}.Invalidate("conn")
	}
	if connID == "" {
		return cerr.ErrFuncArg{}.Invalidate("connID")
	}
	ccp.mu.Lock()
	conn.ReceiveCb(ccp.receiveCb)
	conn.CloseCb(ccp.CallCloseCbs)
	ccp.pool[connID] = conn
	ccp.mu.Unlock()
	return nil
}

// unregister removes User from by uuid of connection of client.
// Used by client as callback when connection closes.
func (ccp *ConnPool) unregister(connID string) {
	ccp.mu.Lock()
	delete(ccp.pool, connID)
	ccp.mu.Unlock()
}

// ReceiveCb sets callback which process messages received by clients
func (ccp *ConnPool) ReceiveCb(cb func(msg []byte, connID string)) {
	ccp.receiveCb = cb
}

// Send func sends message to client by connID of connection.
// Returns ErrWrongConnID error when no any Client
// with such UUID of connection presented in the connpool.
func (ccp *ConnPool) Send(msg []byte, connID string) error {
	ccp.mu.Lock()
	defer ccp.mu.Unlock()
	v, ok := ccp.pool[connID]
	if ok {
		return v.Send(msg)
	}
	return ErrWrongConnID
}

// PingMessageForConn is setter for ping message.
// Will return error when msg is empty string.
func (ccp *ConnPool) PingMessageForConn(msg, connID string) error {
	ccp.mu.Lock()
	defer ccp.mu.Unlock()
	if msg == "" {
		return ErrInvalidPingMsg
	}
	if connID == "" {
		return ErrWrongConnID
	}
	conn, ok := ccp.pool[connID]
	if !ok {
		return ErrWrongConnID
	}
	return conn.PingMessage(msg)
}

// PingWaitForConn is setter for ping timeout.
// Will return error when msg is empty string.
func (ccp *ConnPool) PingWaitForConn(wait int, connID string) error {
	ccp.mu.Lock()
	defer ccp.mu.Unlock()
	if wait < 1 {
		return ErrInvalidPingWait
	}
	if connID == "" {
		return ErrWrongConnID
	}
	conn, ok := ccp.pool[connID]
	if !ok {
		return ErrWrongConnID
	}
	return conn.PingWait(wait)
}
