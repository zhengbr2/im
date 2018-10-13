package main

var (
	round *Round
)

func main() {

	round = NewRound(RoundOptions{
		Timer:     256,
		TimerSize: 2048,
	})

	if err := InitRouterClient([]string{"tcp@localhost:7270"}); err != nil {
		panic(err)
	}

	if err := InitLogicServ([]string{"tcp@localhost:7170"}); err != nil {
		panic(err)
	}

	go InitHttpServ(":7470")

	if err := InitKafka([]string{"localhost:9092"}); err != nil {
		panic(err)
	}

	InitSignal()
}
