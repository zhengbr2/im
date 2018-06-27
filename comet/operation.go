package main

import (
	"im/libs/proto"
	"time"
)

type Operator interface {
	Connect(*proto.Proto, string) (string, int32, time.Duration, error)
	Send(string, *proto.Proto) error
	Disconnect(string) error
}

type DefaultOperator struct {
}

func (operator *DefaultOperator) Connect(p *proto.Proto, rAddr string) (key string, rid int32, heartbeat time.Duration, err error) {
	key, rid, heartbeat, err = connect(p, rAddr)
	return
}

func (operator *DefaultOperator) Send(key string, p *proto.Proto) (err error) {
	err = send(key, p)
	return
}

func (operator *DefaultOperator) Disconnect(key string) (err error) {
	err = disconnect(key)
	return
}
