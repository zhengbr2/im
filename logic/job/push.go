package main

import (
	"encoding/json"
	"im/libs/proto"
	"math/rand"
)

type pushArg struct {
	Key    string
	RoomID int32
	P      *proto.Proto
}

var (
	pushChs    []chan *pushArg
	puchChsLen int
)

func InitPush(puchChan int, size int) {
	pushChs = make([]chan *pushArg, puchChan)
	puchChsLen = puchChan
	for i := 0; i < puchChan; i++ {
		pushChs[i] = make(chan *pushArg, size)
		go processPush(pushChs[i])
	}
}

func processPush(ch chan *pushArg) {
	var arg *pushArg
	for {
		arg = <-ch
		pushRoomKeyMsg(arg.Key, arg.RoomID, arg.P)
	}
}

func push(msg []byte) (err error) {
	m := &proto.KafkaMsg{}
	if err = json.Unmarshal(msg, m); err != nil {
		return
	}

	p := &proto.Proto{}
	p.Operation = m.Operation
	p.Body = m.Body

	pushChs[rand.Int()%puchChsLen] <- &pushArg{Key: m.Key, RoomID: m.RoomId, P: p}
	return
}
