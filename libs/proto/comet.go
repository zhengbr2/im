package proto

type NoArg struct {
}

type NoReply struct {
}

type PushRoomMsgArg struct {
	RoomId int32
	P      *Proto
}

type PushRoomsMsgArg struct {
	RoomIds []int32
	P       *Proto
}

type PushRoomKeyMsgArg struct {
	RoomId int32
	Key    string
	P      *Proto
}
