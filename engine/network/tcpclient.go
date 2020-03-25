package network

import (
	"errors"
	"log"
	"net"
	"sync"
	"time"
)

var (
	ErrClientClosed = errors.New("network: Client closed")
)

type TCPClient struct {
	addr  string
	retry bool

	mutex       sync.RWMutex
	connection  *TCPConnection
	closed      bool
	ConnectChan chan bool
}

func NewTCPClient(addr string) *TCPClient {
	client := &TCPClient{
		addr:        addr,
		retry:       false,
		ConnectChan: make(chan bool),
	}
	return client
}

func (this *TCPClient) EnableRetry()  { this.retry = true }
func (this *TCPClient) DisableRetry() { this.retry = false }

func dialTCP(addr string, timeout time.Duration) (*net.TCPConn, error) {
	if timeout <= 0 {
		raddr, err := net.ResolveTCPAddr("tcp", addr)
		if err != nil {
			return nil, err
		}
		return net.DialTCP("tcp", nil, raddr)
	} else {
		conn, err := net.DialTimeout("tcp", addr, timeout)
		return conn.(*net.TCPConn), err
	}

}
func (this *TCPClient) notifyConnectState(state bool) {
	select {
	case this.ConnectChan <- state:
	default:
	}
}
func (this *TCPClient) DialAndServe(handler TCPHandler, codec TCPCodec, timeout time.Duration) error {
	if this.IsClosed() {
		return ErrClientClosed
	}

	if handler == nil {
		handler = DefaultTCPHandler
	}
	if codec == nil {
		codec = DefaultTCPCodec
	}

	var tempDelay time.Duration // how long to sleep on connect failure
	for {
		conn, err := dialTCP(this.addr, timeout)
		if err != nil {
			if this.IsClosed() {
				this.notifyConnectState(false)
				return ErrClientClosed
			}
			if !this.retry {
				this.notifyConnectState(false)
				return err
			}

			if tempDelay == 0 {
				tempDelay = 1 * time.Second
			} else {
				tempDelay *= 2
			}
			if max := 1 * time.Minute; tempDelay > max {
				tempDelay = max
			}
			log.Printf("TCPClient: dial error: %v; retrying in %v", err, tempDelay)
			time.Sleep(tempDelay)
			continue
		}
		tempDelay = 0

		connection := newTCPConnection(conn)
		if err := this.newConnection(connection); err != nil {
			this.notifyConnectState(false)
			return err
		}
		this.notifyConnectState(true)
		if err := this.serveConnection(connection, handler, codec); err != nil {
			return err
		}
	}
}

func (this *TCPClient) IsClosed() bool {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	return this.closed
}

func (this *TCPClient) newConnection(connection *TCPConnection) error {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.closed {
		return ErrClientClosed
	}
	this.connection = connection
	return nil
}

func (this *TCPClient) serveConnection(connection *TCPConnection, handler TCPHandler, codec TCPCodec) error {
	connection.serve(handler, codec)
	// remove connection
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.closed {
		return ErrClientClosed
	}
	this.connection = nil
	return nil
}

func (this *TCPClient) GetConnection() *TCPConnection {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	if this.closed {
		return nil
	}
	return this.connection
}

func (this *TCPClient) Close() {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if this.closed {
		return
	}
	this.closed = true
	if this.connection == nil {
		return
	}
	this.connection.Close()
	this.connection = nil
}
