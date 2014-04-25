package oracle

import (
	"bufio"
	"errors"
	"log"
	"net"
	"sync"
	"sync/atomic"
)

type Client struct {
	addrport string
	req      chan chan int64
	shutdown int32
	mutex    sync.Mutex
	conn     net.Conn
	reader   *bufio.Reader
	writer   *bufio.Writer
}

func (c *Client) GetTS(num int32) (int64, error) {
	getTs := &GetTS{num}

	c.writer.WriteByte(byte(GETTS))
	getTs.Marshal(c.writer)
	c.writer.Flush()

	msgType, err := c.reader.ReadByte()
	if err != nil {
		return -1, err
	}
	switch uint8(msgType) {
	case REPLYTS:
		replyts := new(ReplyTS)
		if err := replyts.Unmarshal(c.reader); err != nil {
			return -1, err
		}
		return replyts.Timestamp, nil
	default:
		return -1, errors.New("Unknown msg type")
	}
}

func NewClient(address string) (*Client, error) {
	conn_, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}

	cl := &Client{
		addrport: address,
		req:      make(chan chan int64, 100),
		shutdown: 0,
		conn:     conn_,
		writer:   bufio.NewWriter(conn_),
		reader:   bufio.NewReader(conn_),
	}

	go cl.start()
	return cl, nil
}

func (c *Client) start() {
	for atomic.LoadInt32(&c.shutdown) == 0 {
		ch := <-c.req
		l := len(c.req)
		ts, err := c.GetTS(int32(l + 1))
		if err != nil {
			atomic.StoreInt32(&c.shutdown, 1)
			log.Println("get ts error", err)
			c.conn.Close()
			break
		}
		ch <- ts
		for i := 1; i <= l; i++ {
			ch = <-c.req
			ch <- ts - int64(i)
		}
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	select {
	case ch := <-c.req:
		ch <- -1
	default:
		break
	}
}

func (c *Client) Close() {
	atomic.StoreInt32(&c.shutdown, 1)
}

func (c *Client) TS() (int64, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if atomic.LoadInt32(&c.shutdown) == 1 {
		return -1, errors.New("invalid ts")
	}
	ch := make(chan int64)
	c.req <- ch
	if ts := <-ch; ts > -1 {
		return ts, nil
	} else {
		return -1, errors.New("invalid ts")
	}

	// return c.GetTS(1)
}
