package main

import (
	"bufio"
	"im/libs/proto"
)

// 会话管理 （每个连接）
// Channel used by message pusher send msg to write goroutine.
type Channel struct {
	RoomID   int32
	CliProto Ring
	signal   chan *proto.Proto
	Writer   *bufio.Writer
	Reader   *bufio.Reader
}

//cli client ring capacity
//ser signal buffer 
//rid  RoomID
func NewChannel(cli, svr int, rid int32) *Channel {
	c := new(Channel)
	c.RoomID = rid
	c.CliProto.Init(uint64(cli))
	c.signal = make(chan *proto.Proto, svr)
	return c
}

// Push server push message.
func (c *Channel) Push(p *proto.Proto) (err error) {
	select {
	case c.signal <- p:
	default:
	}
	return
}

// Ready check the channel ready or close?
func (c *Channel) Ready() *proto.Proto {
	return <-c.signal
}
// Signal send signal to the channel, protocol ready.
func (c *Channel) Signal() {
	c.signal <- proto.ProtoReady
}

// Close close the channel.
func (c *Channel) Close() {
	c.signal <- proto.ProtoFinish
}
