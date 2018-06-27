package proto

type ConnArg struct {
	P     *Proto
	Sid   int32
	Raddr string
}

type ConnReply struct {
	Key    string
	RoomId int32
}

type SendArg struct {
	Key string
	P   *Proto
}

type SendReply struct {
}

type DisconnArg struct {
	Key string
}

type DisConnReply struct {
}
