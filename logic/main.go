package main

var (
	round *Round
)

func main() {

	round = NewRound(RoundOptions{
		Timer:     256,
		TimerSize: 2048,
	})

	if err := InitRouterRpc([]string{"tcp@localhost:7270"}); err != nil {
		panic(err)
	}

	if err := InitRPC([]string{"tcp@localhost:7170"}); err != nil {
		panic(err)
	}

	go InitHttp(":7470")

	if err := InitKafka([]string{"localhost:9092"}); err != nil {
		panic(err)
	}

	InitSignal()
}
