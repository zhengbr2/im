package main

import (
	inet "im/libs/net"
	"im/libs/net/xrpc"
	"im/libs/proto"
)

var (
	cometRpcClient *xrpc.Clients
	//logicRpcQuit   = make(chan struct{}, 1)
)

const (
	cometService               = "PushRPC"
	cometServicePing           = "PushRPC.Ping"
	cometServicePushRoomKeyMsg = "PushRPC.PushRoomKeyMsg"
)

func InitCometRpc(addrs []string) (err error) {
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

	cometRpcClient = xrpc.Dials(rpcOptions)
	cometRpcClient.Ping(cometServicePing)
	return
}

func pushRoomKeyMsg(key string, roomid int32, p *proto.Proto) (err error) {
	var (
		arg   = proto.PushRoomKeyMsgArg{Key: key, RoomId: roomid, P: p}
		reply = proto.NoReply{}
	)

	err = cometRpcClient.Call(cometServicePushRoomKeyMsg, &arg, &reply)
	return
}
