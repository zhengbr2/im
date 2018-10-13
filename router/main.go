package main

func main() {

	if err := InitRouterServ([]string{"tcp@localhost:7270"}); err != nil {
		panic(err)
	}
	InitSignal()
}
