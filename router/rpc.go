package main

import (
	inet "im/libs/net"
	"im/libs/proto"
	"net"
	"net/rpc"
)

func InitRPC(addrs []string) (err error) {
	var (
		network, addr string
		c             = &RouterRPC{}
	)
	rpc.Register(c)
	for _, bind := range addrs { //[]string{"tcp@localhost:7270"} why cater for multiple net addresses?
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
	defer l.Close()
	rpc.Accept(l)
}

type Key struct {
	Key string `bson:"_id"`
}

type Room struct {
	RoomId int32 `bson:"room_id"`
}

type SerID struct {
	Sid int32 `bson:"s_id"`
}

type channel struct {
	Key   string `bson:"_id"`
	Room  int32  `bson:"room_id"`
	SerID int32  `bson:"s_id"`
	Raddr string `bson:"ip"`
}

type KeyReply struct {
	Room  int32 `bson:"room_id"`
	SerID int32 `bson:"s_id"`
}

type RoomReply struct {
	Key   string `bson:"_id"`
	SerID int32  `bson:"s_id"`
}

type RouterRPC struct {
}

func (r *RouterRPC) Ping(arg *proto.NoArg, reply *proto.NoReply) error {
	return nil
}

func (r *RouterRPC) Put(arg *proto.PutArg, reply *proto.NoReply) (err error) {
	db := DB{}
	defer db.Close()
	err = db.C().Insert(&channel{Key: arg.Key, Room: arg.RoomId, SerID: arg.Sid, Raddr: arg.Raddr})
	return err
}

func (c *RouterRPC) DelAll(arg *proto.NoArg, reply *proto.NoReply) (err error) {
	db := DB{}
	defer db.Close()
	_, err = db.C().RemoveAll(nil)
	return
}

func (c *RouterRPC) DelKey(arg *proto.DelKeyArg, reply *proto.NoReply) (err error) {
	db := DB{}
	defer db.Close()
	_, err = db.C().RemoveAll(&Key{Key: arg.Key})
	return
}

func (c *RouterRPC) DelRoom(arg *proto.DelRoomArg, reply *proto.NoReply) (err error) {
	db := DB{}
	defer db.Close()
	_, err = db.C().RemoveAll(&Room{RoomId: arg.RoomId})
	return
}

func (c *RouterRPC) GetKey(arg *proto.GetKeyArg, reply *proto.GetKeyReply) (err error) {
	result := KeyReply{}
	db := DB{}
	defer db.Close()
	err = db.C().Find(&Key{Key: arg.Key}).One(&result)

	reply.RoomId = result.Room
	reply.Sid = result.SerID

	return
}

func (c *RouterRPC) GetRoom(arg *proto.GetRoomArg, reply *proto.GetRoomReply) (err error) {
	result := RoomReply{}
	db := DB{}
	defer db.Close()
	iter := db.C().Find(&Room{RoomId: arg.RoomId}).Iter()
	for iter.Next(&result) {
		reply.Sid = append(reply.Sid, result.SerID)
		reply.Key = append(reply.Key, result.Key)
	}

	return
}
