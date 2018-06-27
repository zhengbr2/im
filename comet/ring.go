package main

import (
	"im/libs/proto"
)

type Ring struct {
	num  uint64
	rp   uint64
	wp   uint64
	mask uint64
	data []proto.Proto
}

func NewRing(num uint64) *Ring {
	r := new(Ring)
	r.Init(num)
	return r
}

func (r *Ring) Init(num uint64) {
	if num&(num-1) != 0 {
		for num&(num-1) != 0 {
			num &= (num - 1)
		}
		num = num << 1
	}
	r.data = make([]proto.Proto, num)
	r.num = num
	r.mask = r.num - 1
}

func (r *Ring) Get() (proto *proto.Proto, err error) {
	if r.rp == r.wp {
		return nil, ErrRingEmpty
	}
	proto = &r.data[r.rp&r.mask]
	return
}
func (r *Ring) GetAdv() {
	r.rp++
}

func (r *Ring) Set() (proto *proto.Proto, err error) {
	if r.wp-r.rp >= r.num {
		return nil, ErrRingFull
	}
	proto = &r.data[r.wp&r.mask]
	return
}

func (r *Ring) SetAdv() {
	r.wp++
}

func (r *Ring) Reset() {
	r.rp = 0
	r.wp = 0
}
