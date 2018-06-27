package define

const (
	// 心跳
	OP_HEARTBEAT       = uint8(1)
	OP_HEARTBEAT_REPLY = uint8(2)
	// 认证
	OP_AUTH       = uint8(3)
	OP_AUTH_REPLY = uint8(4)
	// 文本消息
	OP_TEXT_SMS       = uint8(5)
	OP_TEXT_SMS_REPLY = uint8(6)
	// proto
	OP_PROTO_READY  = uint8(254)
	OP_PROTO_FINISH = uint8(255)
)
