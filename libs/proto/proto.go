package proto

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"im/libs/define"

	"github.com/gorilla/websocket"
)

const (
	MaxBodySize = uint32(1 << 10)
)

const (
	// size
	HeadSize      = 1
	OperationSize = 1
	BodySize      = 4
	RawHeaderSize = HeadSize + OperationSize + BodySize
	// offset
	HeadOffset      = 0
	OperationOffset = HeadOffset + HeadSize
	BodyOffset      = OperationOffset + OperationSize
)

var (
	emptyProto = Proto{}

	ErrProtoBodyLen = errors.New("default server coder body length error")
)

var (
	ProtoReady  = &Proto{Operation: define.OP_PROTO_READY}
	ProtoFinish = &Proto{Operation: define.OP_PROTO_FINISH}
)

type Proto struct {
	Operation uint8
	Body      []byte
}

func (p *Proto) Reset() {
	*p = emptyProto
}

func (p *Proto) ReadTCP(rr *bufio.Reader) (err error) {
	var (
		//head    uint8
		bodyLen uint32
		buf     []byte
	)
	if buf, err = rr.Peek(RawHeaderSize); err != nil {
		return
	}
	//head 当前应该为0
	//head = buf[HeadOffset]
	p.Operation = buf[OperationOffset]
	bodyLen = binary.BigEndian.Uint32(buf[BodyOffset:])
	if bodyLen > MaxBodySize {
		return ErrProtoBodyLen
	}
	if bodyLen > 0 {
		if buf, err = rr.Peek(int(RawHeaderSize + bodyLen)); err != nil {
			return
		}
		p.Body = buf[RawHeaderSize:]
		_, err = rr.Discard(int(RawHeaderSize + bodyLen))
	} else {
		p.Body = nil
		_, err = rr.Discard(int(RawHeaderSize))
	}
	return
}

func (p *Proto) WriteTCP(wr *bufio.Writer) (err error) {
	var (
		head uint8
	)
	// 固定 0
	head = uint8(0)
	buf := make([]byte, RawHeaderSize)
	buf[HeadOffset] = head
	buf[OperationOffset] = p.Operation
	binary.BigEndian.PutUint32(buf[BodyOffset:], uint32(len(p.Body)))
	if _, err = wr.Write(buf); err != nil {
		return
	}
	if p.Body != nil {
		_, err = wr.Write(p.Body)
	}
	return
}

func (p *Proto) ReadWebsocket(wr *websocket.Conn) (err error) {

	var (
		messageType int //websocket.BinaryMessage
		buf         []byte
	)

	messageType, buf, err = wr.ReadMessage()
	if err != nil {
		return
	}
	if messageType != websocket.BinaryMessage {
		return
	}

	//head 当前应该为0
	//head = buf[HeadOffset]
	p.Operation = buf[OperationOffset]
	bodyLen := binary.BigEndian.Uint32(buf[BodyOffset:])
	if bodyLen > MaxBodySize {
		return ErrProtoBodyLen
	}
	if bodyLen > 0 {
		p.Body = buf[RawHeaderSize:]
	} else {
		p.Body = nil
	}
	return
}

func (p *Proto) WriteWebsocket(wr *websocket.Conn) (err error) {

	if p.Body == nil || len(p.Body) == 0 {
		return
	}

	var (
		head uint8
	)
	// 固定 0
	head = uint8(0)
	buf := make([]byte, RawHeaderSize)
	buf[HeadOffset] = head
	buf[OperationOffset] = p.Operation
	binary.BigEndian.PutUint32(buf[BodyOffset:], uint32(len(p.Body)))

	s := [][]byte{buf, p.Body}
	body := bytes.Join(s, []byte(""))

	err = wr.WriteMessage(websocket.BinaryMessage, body)

	return
}
