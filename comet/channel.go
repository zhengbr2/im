package main

import (
	"bufio"
	"im/libs/proto"
)

// 会话管理 （每个连接）
type Channel struct {
	RoomID   int32
	CliProto Ring
	signal   chan *proto.Proto
	Writer   *bufio.Writer
	Reader   *bufio.Reader
}

func NewChannel(cli, svr int, rid int32) *Channel {
	c := new(Channel)
	c.RoomID = rid
	c.CliProto.Init(uint64(cli))
	c.signal = make(chan *proto.Proto, svr)
	return c
}

func (c *Channel) Push(p *proto.Proto) (err error) {
	select {
	case c.signal <- p:
	default:
	}
	return
}

func (c *Channel) Ready() *proto.Proto {
	return <-c.signal
}

func (c *Channel) Signal() {
	c.signal <- proto.ProtoReady
}

func (c *Channel) Close() {
	c.signal <- proto.ProtoFinish
}
