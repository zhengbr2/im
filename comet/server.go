package main

import (
	"time"
)

var (
	maxInt = 1<<31 - 1
)

type ServerOptions struct {
	CliProto         int
	SvrProto         int
	HandshakeTimeout time.Duration
	TCPRcvbuf        int
	TCPSendBuf       int
}

type Server struct {
	Buckets   []*Bucket
	bucketIdx uint32
	round     *Round
	operator  Operator
	Options   ServerOptions
}

func NewServer(b []*Bucket, r *Round, o Operator, options ServerOptions) *Server {
	s := new(Server)
	s.Buckets = b
	s.bucketIdx = uint32(len(b))
	s.round = r
	s.operator = o
	s.Options = options
	return s
}

func (s *Server) Bucket(subKey int32) *Bucket {
	idx := uint32(subKey) % s.bucketIdx
	return s.Buckets[idx]
}
