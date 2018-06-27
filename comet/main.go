package main

import (
	"runtime"
	"time"
)

var (
	DefaultServer *Server
)

func main() {

	if err := InitLogicRpc([]string{"tcp@localhost:7170"}); err != nil {
		panic(err)
	}

	buckets := make([]*Bucket, 1024)
	for i := 0; i < 1024; i++ {
		buckets[i] = NewBucket(BucketOptions{
			ChannelSize: 1024,
		})
	}
	round := NewRound(RoundOptions{
		Timer:     256,
		TimerSize: 2048,
	})

	operator := new(DefaultOperator)
	DefaultServer = NewServer(buckets, round, operator, ServerOptions{
		CliProto:         5,
		SvrProto:         80,
		HandshakeTimeout: 5 * time.Second,
		TCPRcvbuf:        256,
		TCPSendBuf:       2048,
	})

	if err := InitTCP([]string{":8081"}, runtime.NumCPU()); err != nil {
		panic(err)
	}

	if err := InitWebsocket([]string{":8082"}); err != nil {
		panic(err)
	}

	if err := InitRPCPuch([]string{"tcp@localhost:8092"}); err != nil {
		panic(err)
	}

	InitSignal()
}
