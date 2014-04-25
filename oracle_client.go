package oracle

import (
	"bufio"
	"errors"
	"net"
	"sync"
	"sync/atomic"
)

type Client struct {
	addrport string
	shutdown int32
	mutex    sync.Mutex
	conn     net.Conn
	reader   *bufio.Reader
	writer   *bufio.Writer
}

func (c *Client) GetTS(num int32) (int64, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if atomic.LoadInt32(&c.shutdown) == 1 {
		return -1, errors.New("already close")
	}

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
		shutdown: 0,
		conn:     conn_,
		writer:   bufio.NewWriter(conn_),
		reader:   bufio.NewReader(conn_),
	}
	return cl, nil
}

func (c *Client) Close() {
	atomic.StoreInt32(&c.shutdown, 1)
}

func (c *Client) TS() (int64, error) {
	return c.GetTS(1)
}
