package oracle

import (
	"bufio"
	"encoding/binary"
	"hash/crc32"
	"io"
	"log"
	"net"
	"os"
	"sync"
)

type Oracle struct {
	maxTs     int64
	remain    int32
	batchsize int32
	addrport  string
	shutdown  bool
	mutex     sync.Mutex
	bookeeper *os.File
}

func NewOracle() *Oracle {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	bk, err := os.OpenFile("orc.log", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalln("Cannot open log file")
	}
	return &Oracle{
		maxTs:     -1,
		remain:    0,
		batchsize: 10000,
		addrport:  ":7070",
		shutdown:  false,
		bookeeper: bk,
	}
}

func (o *Oracle) WaitForClientConnections() {
	listener, err := net.Listen("tcp", o.addrport)
	if err != nil {
		log.Panicln("Listen error", err)
	}
	defer listener.Close()
	for !o.shutdown {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Accept error", err)
			continue
		}
		go o.ServeConn(conn)
	}
}

func (o *Oracle) Close() {
	o.shutdown = true
}

func (o *Oracle) ServeConn(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)
	for {
		msgType, err := reader.ReadByte()
		if err != nil {
			return
		}
		switch uint8(msgType) {
		case GETTS:
			getts := new(GetTS)
			if err := getts.Unmarshal(reader); err != nil {
				return
			}
			replyts := o.getTimestamp(getts)
			writer.WriteByte(byte(REPLYTS))
			replyts.Marshal(writer)
			writer.Flush()
		default:
			return
		}
	}
}

func (o *Oracle) getTimestamp(getTs *GetTS) *ReplyTS {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	if getTs.Num > o.remain {
		batch := o.batchsize
		if batch < getTs.Num {
			batch = getTs.Num
		}
		// allocate batch ts
		o.maxTs += int64(batch)
		o.remain += batch
		o.log(o.maxTs)
	}

	o.remain -= getTs.Num
	return &ReplyTS{o.maxTs - int64(o.remain)}
}

func (o *Oracle) Recover() {
	r, err := o.bookeeper.Seek(0, os.SEEK_END)
	if err != nil {
		log.Panicln("log seek error")
	}

	if rr := r % 12; rr != 0 {
		r -= rr
		if err := o.bookeeper.Truncate(r); err != nil {
			log.Panicln("log truncate error")
		}
	}

	if r <= 0 {
		o.maxTs = -1
		o.remain = 0
		log.Println("maxts = ", o.maxTs)
		return
	}
	// read last log record
	if _, err := o.bookeeper.Seek(-12, os.SEEK_CUR); err != nil {
		log.Panicln("log seek error")
	}
	var b [12]byte
	cs := b[:4]
	bs := b[4:12]
	if _, err := io.ReadFull(o.bookeeper, b[:12]); err != nil {
		log.Panicln("read log error")
	}
	crc := crc32.ChecksumIEEE(bs)
	if crc != binary.LittleEndian.Uint32(cs) {
		log.Panicln("crc error")
	}
	ts := int64(binary.LittleEndian.Uint64(bs))
	o.maxTs = ts
	o.remain = 0
	log.Println("maxts = ", o.maxTs)
	return
}

func (o *Oracle) log(ts int64) {
	// log.Println("log...", ts)
	var b [12]byte
	bs := b[4:12]
	cs := b[:4]
	binary.LittleEndian.PutUint64(bs, uint64(ts))
	crc := crc32.ChecksumIEEE(bs)
	binary.LittleEndian.PutUint32(cs, crc)
	o.bookeeper.Write(b[:12])
}
