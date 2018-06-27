package main

import (
	inet "im/libs/net"
	"im/libs/net/xrpc"
	"im/libs/proto"
	"time"
)

var (
	logicRpcClient *xrpc.Clients
	//logicRpcQuit   = make(chan struct{}, 1)
)

const (
	logicService           = "RPC"
	logicServicePing       = "RPC.Ping"
	logicServiceConnect    = "RPC.Connect"
	logicServiceSend       = "RPC.Send"
	logicServiceDisconnect = "RPC.Disconnect"
	logicServiceDelAll     = "RPC.DelAll"
)

func InitLogicRpc(addrs []string) (err error) {
	var (
		bind          string
		network, addr string
		rpcOptions    []xrpc.ClientOptions
	)

	for _, bind = range addrs {
		if network, addr, err = inet.ParseNetwork(bind); err != nil {
			return
		}
		options := xrpc.ClientOptions{
			Proto: network,
			Addr:  addr,
		}
		rpcOptions = append(rpcOptions, options)
	}

	logicRpcClient = xrpc.Dials(rpcOptions)
	logicRpcClient.Ping(logicServicePing)
	resetRouter()
	return
}

func connect(p *proto.Proto, rAddr string) (key string, rid int32, heartbeat time.Duration, err error) {
	var (
		arg   = proto.ConnArg{P: p, Sid: 1, Raddr: rAddr}
		reply = proto.ConnReply{}
	)
	if err = logicRpcClient.Call(logicServiceConnect, &arg, &reply); err != nil {
		return
	}
	key = reply.Key
	rid = reply.RoomId
	heartbeat = 5 * time.Minute
	return
}

func send(key string, p *proto.Proto) (err error) {
	var (
		arg   = proto.SendArg{Key: key, P: p}
		reply = proto.SendReply{}
	)
	err = logicRpcClient.Call(logicServiceSend, &arg, &reply)
	return
}

func disconnect(key string) (err error) {
	var (
		arg   = proto.DisconnArg{Key: key}
		reply = proto.DisConnReply{}
	)
	err = logicRpcClient.Call(logicServiceDisconnect, &arg, &reply)
	return
}

func resetRouter() (err error) {
	err = logicRpcClient.Call(logicServiceDelAll, nil, nil)
	return
}
