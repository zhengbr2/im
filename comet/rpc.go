package main

import (
	inet "im/libs/net"
	"im/libs/proto"
	"net"
	"net/rpc"
)

func InitRPCPuch(addrs []string) (err error) {
	var (
		bind          string
		network, addr string
		c             = &PushRPC{}
	)
	rpc.Register(c)
	for _, bind = range addrs {
		if network, addr, err = inet.ParseNetwork(bind); err != nil {
			return
		}
		go rpcListen(network, addr)
	}
	return
}

func rpcListen(network, addr string) {
	l, err := net.Listen(network, addr)
	if err != nil {
		panic(err)
	}
	defer func() {
		l.Close()
	}()
	rpc.Accept(l)
}

type PushRPC struct {
}

func (this *PushRPC) Ping(arg *proto.NoArg, reply *proto.NoReply) error {
	return nil
}

func (this *PushRPC) PushRoomKeyMsg(arg *proto.PushRoomKeyMsgArg, reply *proto.NoReply) (err error) {
	var (
		bucket  *Bucket
		channel *Channel
	)
	bucket = DefaultServer.Bucket(arg.RoomId)
	if channel = bucket.Channel(arg.Key); channel != nil {
		err = channel.Push(arg.P)
	}
	return
}
