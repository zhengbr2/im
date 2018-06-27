package main

func main() {

	if err := InitRPC([]string{"tcp@localhost:7270"}); err != nil {
		panic(err)
	}
	InitSignal()
}
