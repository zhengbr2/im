package main

func main() {

	if err := InitCometRpc([]string{"tcp@localhost:8092"}); err != nil {
		panic(err)
	}

	InitPush(16, 100)

	if err := InitKafka([]string{"localhost:9092"}); err != nil {
		panic(err)
	}

	InitSignal()
}
