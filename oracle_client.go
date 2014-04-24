package oracle

import (
	"bufio"
	"errors"
	"net"
)

type Client struct {
	Addrport string
}

func (c *Client) GetTS(num int32) (int64, error) {
	getTs := &GetTS{num}
	conn, err := net.Dial("tcp", c.Addrport)
	if err != nil {
		return -1, err
	}
	defer conn.Close()

	writer := bufio.NewWriter(conn)
	reader := bufio.NewReader(conn)
	writer.WriteByte(byte(GETTS))
	getTs.Marshal(writer)
	writer.Flush()

	msgType, err := reader.ReadByte()
	if err != nil {
		return -1, err
	}
	switch uint8(msgType) {
	case REPLYTS:
		replyts := new(ReplyTS)
		if err := replyts.Unmarshal(reader); err != nil {
			return -1, err
		}
		return replyts.Timestamp, nil
	default:
		return -1, errors.New("Unknown msg type")
	}
}
