package main

import (
	"errors"
	"im/libs/cityhash"
	"im/libs/define"
	inet "im/libs/net"
	"im/libs/proto"
	itime "im/libs/time"
	"net"
	"net/rpc"
	"strconv"
	"sync"
	"time"

	"github.com/globalsign/mgo/bson"
	goproto "github.com/golang/protobuf/proto"
)

var (
	replyTimer     *ReplyTimer
	pushReplayTime = time.Second * 10
)

type ReplyTimer struct {
	Data map[uint64]*itime.TimerData
	Lock sync.Mutex
}

func InitReplyTimer() (reply *ReplyTimer) {
	reply = &ReplyTimer{}
	reply.Data = make(map[uint64]*itime.TimerData)
	return
}

func (r *ReplyTimer) Get(k uint64) *itime.TimerData {
	r.Lock.Lock()
	defer r.Lock.Unlock()
	return r.Data[k]
}
func (r *ReplyTimer) Set(k uint64, v *itime.TimerData) {
	r.Lock.Lock()
	defer r.Lock.Unlock()
	r.Data[k] = v
}
func (r *ReplyTimer) Remove(k uint64) {
	r.Lock.Lock()
	defer r.Lock.Unlock()
	delete(r.Data, k)
}

func InitLogicServ(addrs []string) (err error) {
	var (
		network, addr string
		c             = &RPC{}
	)
	rpc.Register(c)
	for _, bind := range addrs {
		if network, addr, err = inet.ParseNetwork(bind); err != nil {
			return
		}
		go rpcListen(network, addr)
	}
	replyTimer = InitReplyTimer()
	return
}

func rpcListen(network, addr string) {
	l, err := net.Listen(network, addr)
	if err != nil {
		panic(err)
	}
	defer l.Close()
	rpc.Accept(l)
}

type RPC struct {
}

func (r *RPC) Ping(arg *proto.NoArg, reply *proto.NoReply) error {
	return nil
}

type User struct {
	Key string `bson:"key"`
}

func (r *RPC) Connect(arg *proto.ConnArg, reply *proto.ConnReply) (err error) {

	auth := &proto.Auth{}
	err = goproto.Unmarshal(arg.P.Body, auth)
	if err != nil {
		return err
	}

	uid := auth.GetUid()
	key := auth.GetKey()

	user := User{}
	db := DB{}
	err = db.C("users").Find(bson.M{"_id": uid}).One(&user)
	defer db.Close()
	if err != nil {
		return err
	}

	if user.Key == "" || user.Key != key {
		return errors.New("error auth -> key")
	}

	key = bson.NewObjectId().Hex()

	reply.Key = key
	reply.RoomId = uid

	putarg := proto.PutArg{RoomId: uid, Sid: arg.Sid, Key: key, Raddr: arg.Raddr}
	err = put(&putarg)
	return
}

type FindUidsToMid struct {
	Uids []int32 `bson:"uids"`
}
type FindFromToMid struct {
	Id string `bson:"_id"`
}

type PushMsg struct {
	Time        string  `bson:"time"`
	Sid         string  `bson:"_id"`
	From        int32   `bson:"fromUid"`
	FromMid     string  `bson:"fromMid"`
	ToUids      []int32 `bson:"toUids"`
	Type        string  `bson:"type"`
	ContentType string  `bson:"contentType"`
	Content     string  `bson:"content"`
	Size        int32   `bson:"size"`
	W           int32   `bson:"w"`
	H           int32   `bson:"h"`
	ReadUids    []int32 `bson:"readUids"`
}

func (c *RPC) Send(arg *proto.SendArg, reply *proto.SendReply) (err error) {

	var (
		getKeyArg   = &proto.GetKeyArg{Key: arg.Key}
		getKeyReply = &proto.GetKeyReply{}
	)
	err = getKey(getKeyArg, getKeyReply)
	if err != nil {
		return
	}

	msg := &proto.Msg{}
	err = goproto.Unmarshal(arg.P.Body, msg)
	if err != nil {
		return err
	}

	go sendHandle(arg, getKeyReply.RoomId, msg)

	return
}

func sendHandle(arg *proto.SendArg, keyRoomId int32, msg *proto.Msg) (err error) {

	db := DB{}
	defer db.Close()

	if arg.P.Operation == define.OP_TEXT_SMS {

		if msg.GetTo() == "" {
			return errors.New("proto error -> to is nil")
		}

		msg.From = keyRoomId

		msg.Sid = bson.NewObjectId().Hex()

		msg.Time = strconv.FormatInt(time.Now().Unix(), 10)

		var fromMid string
		isFromMid := true
		// 查询来源是否在目标房间
		mid := FindFromToMid{}
		err = db.C("mids").Find(bson.M{
			"_id":  msg.To,
			"uids": bson.M{"$in": msg.From},
		}).One(&mid)
		if err != nil || mid.Id == "" {
			isFromMid = false
			// 来源不在目标房间，查询来源房间号
			from := []int32{msg.From}
			err = db.C("mids").Find(bson.M{
				"uids": from,
			}).One(&mid)
			if err != nil || mid.Id == "" {
				// 未找到来源房间号
				return
			}
		}
		fromMid = mid.Id

		// 获取当前目标房间接收人
		uids := FindUidsToMid{}
		err = db.C("mids").Find(bson.M{"_id": msg.To}).One(&uids)
		isOldPush := false
		if err != nil || len(uids.Uids) == 0 {
			// 未找到目标房间,推送到其他业务处理（该im与主业务分离）
			isOldPush = true
		}

		pushMsg := PushMsg{
			Time:        msg.Time,
			Sid:         msg.Sid,
			From:        msg.From,
			FromMid:     fromMid,
			ToUids:      uids.Uids,
			Type:        msg.GetType(),
			ContentType: msg.GetContentType(),
			Content:     msg.GetContent(),
			Size:        msg.GetSize(),
			W:           msg.GetW(),
			H:           msg.GetH(),
		}
		err = db.C("msgs").Insert(&pushMsg)
		if err != nil {
			return
		}

		if arg.Key != "" {
			// 回执来源
			replyMsg := proto.Proto{}
			replyMsg.Operation = define.OP_TEXT_SMS_REPLY
			replyMsg.Body, err = goproto.Marshal(
				&proto.Msg{
					Sid:  msg.Sid,
					Id:   msg.Id,
					Time: msg.Time,
				})
			if err != nil {
				return
			}
			err = mpushKafka(arg.Key, keyRoomId, &replyMsg)
			if err != nil {
				return
			}
		}

		pushData := &PushData{
			From:        pushMsg.From,
			To:          msg.To,
			ContentType: pushMsg.ContentType,
			Content:     pushMsg.Content,
		}

		protoMsg := proto.Proto{}
		protoMsg.Operation = define.OP_TEXT_SMS
		protoMsg.Body, err = goproto.Marshal(msg)
		if err != nil {
			return
		}
		if isOldPush {
			go PostPushData(pushData)
		}

		tomsg := &proto.Msg{
			Time:        msg.Time,
			Sid:         msg.Sid,
			From:        msg.From,
			To:          fromMid,
			Type:        msg.Type,
			ContentType: msg.ContentType,
			Content:     msg.Content,
			Size:        msg.Size,
			W:           msg.W,
			H:           msg.H,
		}
		toProtoMsg := proto.Proto{}
		toProtoMsg.Operation = define.OP_TEXT_SMS
		toProtoMsg.Body, err = goproto.Marshal(tomsg)
		if err != nil {
			return
		}

		if isFromMid {
			// 如果来源在目标房间
			// 目标房间所有用户 uids.Uids
			for _, value := range uids.Uids {
				getRoomArg := &proto.GetRoomArg{RoomId: value}
				getRoomReply := &proto.GetRoomReply{}
				err = getRoom(getRoomArg, getRoomReply)
				if err != nil || len(getRoomReply.Key) <= 0 {
					// 用户离线 推送
					pushData.To = strconv.Itoa(int(value))
					go PostPushData(pushData)
				} else {
					for _, key := range getRoomReply.Key {
						if key != arg.Key {
							if value == msg.From {
								timeMPush(msg.Sid, key, value, &protoMsg)
							} else {
								timeMPush(msg.Sid, key, value, &toProtoMsg)
							}
						}
					}
				}
			}
		} else {
			// 来源不在目标房间 来源多端同步
			uids = FindUidsToMid{}
			err = db.C("mids").Find(bson.M{"_id": fromMid}).One(&uids)
			for _, value := range uids.Uids {
				getRoomArg := &proto.GetRoomArg{RoomId: value}
				getRoomReply := &proto.GetRoomReply{}
				err = getRoom(getRoomArg, getRoomReply)
				if err == nil {
					for _, key := range getRoomReply.Key {
						if key != arg.Key {
							timeMPush(msg.Sid, key, value, &protoMsg)
						}
					}
				}
			}

			// 发送目标房间
			uids = FindUidsToMid{}
			err = db.C("mids").Find(bson.M{"_id": msg.To}).One(&uids)
			for _, value := range uids.Uids {
				getRoomArg := &proto.GetRoomArg{RoomId: value}
				getRoomReply := &proto.GetRoomReply{}
				err = getRoom(getRoomArg, getRoomReply)
				if err != nil || len(getRoomReply.Key) <= 0 {
					// 用户离线 推送
					pushData.To = strconv.Itoa(int(value))
					go PostPushData(pushData)
				} else {
					for _, key := range getRoomReply.Key {
						if key != arg.Key {
							timeMPush(msg.Sid, key, value, &toProtoMsg)
						}
					}
				}
			}
		}

	} else if arg.P.Operation == define.OP_TEXT_SMS_REPLY || arg.P.Operation == define.OP_TEXT_SMS_WEB_REPLY {

		if arg.P.Operation == define.OP_TEXT_SMS_REPLY {
			selector := bson.M{"_id": msg.Sid}
			data := bson.M{"$addToSet": bson.M{"readUids": keyRoomId}}
			err = db.C("msgs").Update(selector, data)
		}

		subKey := arg.Key + msg.Sid
		tr := round.TimerWithKey(subKey)
		r := cityhash.CityHash64([]byte(subKey), uint32(len(subKey)))
		if trd := replyTimer.Get(r); trd != nil {
			tr.Del(trd)
		}
		replyTimer.Remove(r)
	}
	return
}

func timeMPush(sid string, key string, value int32, protoMsg *proto.Proto) (err error) {
	err = mpushKafka(key, value, protoMsg)

	subKey := key + sid
	tr := round.TimerWithKey(subKey)
	var count time.Duration
	count = 1
	r := cityhash.CityHash64([]byte(subKey), uint32(len(subKey)))
	var trd *itime.TimerData
	trd = tr.Add(pushReplayTime, func() {
		mpushKafka(key, value, protoMsg)
		count++
		tr.Set(trd, pushReplayTime*count)
		if count >= 3 {
			tr.Del(trd)
			replyTimer.Remove(r)
		}
	})
	replyTimer.Set(r, trd)
	return
}

func (r *RPC) Disconnect(arg *proto.DisconnArg, reply *proto.DisConnReply) (err error) {
	delarg := proto.DelKeyArg{Key: arg.Key}
	err = delKey(&delarg)
	return
}

func (r *RPC) DelAll(arg *proto.NoArg, reply *proto.NoReply) (err error) {
	err = delAll()
	return
}
