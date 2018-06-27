package proto

type KafkaMsg struct {
	Key       string `json:"key"`
	RoomId    int32  `json:roomid`
	Operation uint8  `json:"op"`
	Body      []byte `json:"body"`
}
