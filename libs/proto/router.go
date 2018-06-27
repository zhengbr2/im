package proto

type PutArg struct {
	RoomId int32
	Sid    int32
	Key    string
	Raddr  string
}

type DelKeyArg struct {
	Key string
}

type DelRoomArg struct {
	RoomId int32
}

type GetKeyArg struct {
	Key string
}

type GetKeyReply struct {
	RoomId int32
	Sid    int32
}

type GetRoomArg struct {
	RoomId int32
}

type GetRoomReply struct {
	Key []string
	Sid []int32
}
