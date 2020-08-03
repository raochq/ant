package network

import (
	"bufio"
	"log"
	"net"
	"runtime"
	"sync"
)

type TCPConnect interface {
	net.Conn
	Shutdown()
	Send(...[]byte) bool
}

type TCPHandler interface {
	Connect(TCPConnect)
	Disconnect(TCPConnect)
	Receive(TCPConnect, []byte)
}

type TCPConnection struct {
	*net.TCPConn

	bufs   [][]byte
	mutex  sync.Mutex
	cond   *sync.Cond
	closed bool
}

type defaultTCPHandler struct {
}

func (*defaultTCPHandler) Connect(TCPConnect) {

}

func (*defaultTCPHandler) Disconnect(TCPConnect) {

}

func (*defaultTCPHandler) Receive(TCPConnect, []byte) {

}

var DefaultTCPHandler = &defaultTCPHandler{}

func newTCPConnection(conn *net.TCPConn) *TCPConnection {
	connection := &TCPConnection{
		TCPConn: conn,
	}
	connection.cond = sync.NewCond(&connection.mutex)
	conn.SetNoDelay(true) // no delay

	return connection
}

func (this *TCPConnection) serve(handler TCPHandler, codec TCPCodec) {
	defer func() {
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			log.Printf("network: panic serving %v: %v\n%s", this.RemoteAddr(), err, buf)
		}
		this.TCPConn.Close()
	}()
	if handler == nil {
		handler = DefaultTCPHandler
	}
	//this.handler = handler

	this.startBackgroundWrite(codec)
	defer this.stopBackgroundWrite()

	handler.Connect(this)
	defer handler.Disconnect(this)

	// loop read
	r := bufio.NewReader(this.TCPConn)
	for {
		b, err := codec.Read(r)
		if err != nil {
			return
		}
		handler.Receive(this, b)
	}
}

func (this *TCPConnection) startBackgroundWrite(codec TCPCodec) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.closed {
		return
	}
	go this.backgroundWrite(codec)
}

func (this *TCPConnection) backgroundWrite(codec TCPCodec) {
	defer func() {
		if err := recover(); err != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			log.Printf("network: panic serving %v: %v\n%s", this.RemoteAddr(), err, buf)
		}
		this.TCPConn.Close()
	}()

	// loop write
	w := bufio.NewWriter(this.TCPConn)
	for closed := false; !closed; {
		var bufs [][]byte

		this.mutex.Lock()
		for !this.closed && len(this.bufs) == 0 {
			this.cond.Wait()
		}
		bufs, this.bufs = this.bufs, bufs // swap
		closed = this.closed
		this.mutex.Unlock()

		for _, b := range bufs {
			if err := codec.Write(w, b); err != nil {
				this.closeSend()
				return
			}
		}
		if err := w.Flush(); err != nil {
			this.closeSend()
			return
		}
	}
}

func (this *TCPConnection) stopBackgroundWrite() {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.closed {
		return
	}
	this.closed = true
	this.cond.Signal()
}

func (this *TCPConnection) closeSend() {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.closed {
		return
	}
	this.closed = true
}
func (this *TCPConnection) Send(b ...[]byte) bool {
	this.mutex.Lock()
	if this.closed {
		this.mutex.Unlock()
		return false
	}
	this.bufs = append(this.bufs, b...)
	this.mutex.Unlock()

	this.cond.Signal()
	return true
}

func (this *TCPConnection) ForceClose() {
	this.TCPConn.SetLinger(0)
	this.TCPConn.Close()
}

func (this *TCPConnection) Shutdown() {
	this.stopBackgroundWrite() // stop write
}
