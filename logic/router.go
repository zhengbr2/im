package main

import (
	inet "im/libs/net"
	"im/libs/net/xrpc"
	"im/libs/proto"
)

var (
	routerRpcClient *xrpc.Clients
	//logicRpcQuit   = make(chan struct{}, 1)
)

const (
	routerService        = "RouterRPC"
	routerServicePing    = "RouterRPC.Ping"
	routerServicePut     = "RouterRPC.Put"
	routerServiceDelAll  = "RouterRPC.DelAll"
	routerServiceDelRoom = "RouterRPC.DelRoom"
	routerServiceDelKey  = "RouterRPC.DelKey"
	routerServiceGetKey  = "RouterRPC.GetKey"
	routerServiceGetRoom = "RouterRPC.GetRoom"
)

func InitRouterRpc(addrs []string) (err error) {
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

	routerRpcClient = xrpc.Dials(rpcOptions)
	routerRpcClient.Ping(routerServicePing)
	return
}

func put(arg *proto.PutArg) (err error) {
	err = routerRpcClient.Call(routerServicePut, arg, nil)
	return
}

func delAll() (err error) {
	err = routerRpcClient.Call(routerServiceDelAll, nil, nil)
	return
}

func delRoom(arg *proto.DelRoomArg) (err error) {
	err = routerRpcClient.Call(routerServiceDelRoom, arg, nil)
	return
}

func delKey(arg *proto.DelKeyArg) (err error) {
	err = routerRpcClient.Call(routerServiceDelKey, arg, nil)
	return

}

func getKey(arg *proto.GetKeyArg, reply *proto.GetKeyReply) (err error) {
	err = routerRpcClient.Call(routerServiceGetKey, arg, reply)
	return
}

func getRoom(arg *proto.GetRoomArg, reply *proto.GetRoomReply) (err error) {
	err = routerRpcClient.Call(routerServiceGetRoom, arg, reply)
	return
}
