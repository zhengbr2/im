package main

import (
	"bytes"
	"im/libs/proto"
	"net/http"

	"encoding/json"
	"im/libs/define"

	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo/bson"
)

func InitHttpServ(addr string) {

	//gin.DisableConsoleColor()
	gin.SetMode(gin.ReleaseMode)
	//router := gin.Default()
	router := gin.New()

	im := router.Group("/im")
	{
		im.POST("/register", register)
		im.POST("/send", send)
		im.POST("/msg", getmsg)
		im.POST("/getmsgcount", getmsgcount)
		im.POST("/readmsg", readmsg)
		im.POST("/useronline", useronline)
	}

	router.Run(addr)
}

type RegisterParame struct {
	Uid  int32   `json:"uid"`
	Mid  string  `json:"mid"`
	Uids []int32 `json:"uids"`
	Type string  `json:"type"` //在mid存在并且uids>0 的情况下 1增加 2移除
}

type SendParame struct {
	From        int32    `json:"from"`
	To          []string `json:"to"`
	Type        string   `json:"type"`
	ContentType string   `json:"contentType"`
	Content     string   `json:"content"`
	Size        int32    `json:"size"`
	W           int32    `json:"w"`
	H           int32    `json:"h"`
}

type GetMsgParame struct {
	Uid  int32  `json:"uid"`
	Time string `json:"time"` //未读消息正序 历史消息倒序
	Mid  string `json:"mid"`  //没有mid 为未读消息 有mid 为历史消息
	Size int32  `json:"size"`
}

type ReadMsgParame struct {
	Uid int32    `json:"uid"`
	Sid []string `json:"sid"`
}

type UseronlineParame struct {
	Uid int32 `json:"uid"`
}

func register(c *gin.Context) {

	var (
		parame RegisterParame
		key    string
	)

	db := DB{}
	defer db.Close()

	err := c.BindJSON(&parame)

	if err == nil && parame.Uid != 0 {

		selector := bson.M{"_id": parame.Uid}
		key = bson.NewObjectId().Hex()
		data := bson.M{"$set": bson.M{"key": key}}
		_, err = db.C("users").Upsert(selector, data)
	}

	if err == nil && parame.Mid != "" {
		if len(parame.Uids) > 0 {
			selector := bson.M{"_id": parame.Mid}
			var data bson.M
			if parame.Type == "" {
				data = bson.M{"$set": bson.M{"uids": parame.Uids}}
				_, err = db.C("mids").Upsert(selector, data)
			} else if parame.Type == "1" {
				data = bson.M{"$addToSet": bson.M{"uids": bson.M{"$each": parame.Uids}}}
				err = db.C("mids").Update(selector, data)
			} else if parame.Type == "2" {
				data = bson.M{"$pull": bson.M{"uids": bson.M{"$in": parame.Uids}}}
				err = db.C("mids").Update(selector, data)
			}
		} else if parame.Type == "2" {
			err = db.C("mids").Remove(bson.M{"_id": parame.Mid})
		}
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"key":    key,
	})
}

type FindToMsg struct {
	Time        string `bson:"time" json:"time"`
	Sid         string `bson:"_id" json:"sid"`
	From        int32  `bson:"fromUid" json:"from"`
	To          string `bson:"fromMid" json:"to"`
	Type        string `bson:"type" json:"type"`
	ContentType string `bson:"contentType" json:"contentType"`
	Content     string `bson:"content" json:"content"`
	Size        int32  `bson:"size" json:"size"`
	W           int32  `bson:"w" json:"w"`
	H           int32  `bson:"h" json:"h"`
}

func getmsgcount(c *gin.Context) {
	var (
		parame GetMsgParame
	)

	err := c.BindJSON(&parame)

	if parame.Uid == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "parame error",
		})
		return
	}

	db := DB{}
	defer db.Close()

	var a [1]int32
	a[0] = parame.Uid
	count := 0
	count, err = db.C("msgs").Find(bson.M{
		"toUids":   parame.Uid,
		"readUids": bson.M{"$nin": a},
		"time":     bson.M{"$gte": parame.Time},
	}).Count()

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "no Message",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"count":  count,
	})
}

func getmsg(c *gin.Context) {
	var (
		parame GetMsgParame
	)

	err := c.BindJSON(&parame)

	if parame.Uid == 0 || parame.Size <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "parame error",
		})
		return
	}

	db := DB{}
	defer db.Close()
	var findToMsg []FindToMsg
	if parame.Mid == "" {
		var a [1]int32
		a[0] = parame.Uid
		err = db.C("msgs").Find(bson.M{
			"toUids":   parame.Uid,
			"readUids": bson.M{"$nin": a},
			"time":     bson.M{"$gte": parame.Time},
		}).Sort("time").Limit(int(parame.Size)).Iter().All(&findToMsg)
	} else {
		if err == nil {
			type FindOneMid struct {
				Uids []int32 `bson:"uids"`
			}
			findOneMid := FindOneMid{}

			err = db.C("mids").Find(bson.M{"_id": parame.Mid}).One(&findOneMid)

			if err == nil && len(findOneMid.Uids) == 1 {
				err = db.C("msgs").Find(bson.M{
					"$or": []bson.M{
						bson.M{
							"fromMid": parame.Mid,
							"toUids":  parame.Uid},
						bson.M{
							"fromUid": parame.Uid,
							"toUids":  findOneMid.Uids}},
					"time": bson.M{"$lte": parame.Time},
				}).Sort("-time").Limit(int(parame.Size)).Iter().All(&findToMsg)
			} else {
				err = db.C("msgs").Find(bson.M{
					"fromMid": parame.Mid,
					"toUids":  parame.Uid,
					"time":    bson.M{"$lte": parame.Time},
				}).Sort("-time").Limit(int(parame.Size)).Iter().All(&findToMsg)
			}
		}
	}

	if err != nil || len(findToMsg) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "no Message",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"list":   findToMsg,
	})
}

func readmsg(c *gin.Context) {
	var (
		parame ReadMsgParame
	)

	err := c.BindJSON(&parame)

	if err == nil && parame.Uid > 0 && len(parame.Sid) > 0 {
		go readHandle(&parame)
	}
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "parame error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}

func readHandle(parame *ReadMsgParame) {
	db := DB{}
	defer db.Close()

	for _, sid := range parame.Sid {
		selector := bson.M{"_id": sid}
		data := bson.M{"$addToSet": bson.M{"readUids": parame.Uid}}
		db.C("msgs").Update(selector, data)
	}
}

type FindUseronline struct {
	Key string `bson:"_id"`
}

func useronline(c *gin.Context) {
	var (
		parame UseronlineParame
	)

	err := c.BindJSON(&parame)

	if err == nil && parame.Uid != 0 {

		db := DB{}
		defer db.Close()

		selector := bson.M{"room_id": parame.Uid}
		uids := FindUseronline{}
		err = db.C("router").Find(selector).One(&uids)
		if err == nil && uids.Key != "" {
			c.JSON(http.StatusOK, gin.H{
				"status": "success",
				"online": "1",
			})
			return
		}
	}

	c.JSON(http.StatusBadRequest, gin.H{
		"status":  "error",
		"message": "parame error",
	})
}

func send(c *gin.Context) {

	var (
		parame SendParame
	)

	err := c.BindJSON(&parame)

	if err == nil && parame.From > 0 && len(parame.To) > 0 {

		go pushHandle(&parame)
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  "error",
			"message": "parame error",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status": "success",
	})
}

func pushHandle(parame *SendParame) {
	for _, to := range parame.To {
		msg := &proto.Msg{}
		msg.From = parame.From
		msg.To = to
		msg.ContentType = parame.ContentType
		msg.Content = parame.Content
		if parame.Size > 0 {
			msg.Size = parame.Size
			msg.W = parame.W
			msg.H = parame.H
		}
		arg := &proto.SendArg{Key: "", P: &proto.Proto{Operation: define.OP_TEXT_SMS}}
		go sendHandle(arg, parame.From, msg)
	}
}

type PushData struct {
	From        int32  `json:"from"`
	To          string `json:"to"`
	ContentType string `json:"contentType"`
	Content     string `json:"content"`
}

// 用户离线，推送消息api
// 此api无效，改为你的推送api
var postPushDataUrl = "http://192.168.1.100:8099/api/IM/PushMsg"

func PostPushData(data *PushData) {

	body, err := json.Marshal(data)
	if err != nil {
		return
	}
	reader := bytes.NewReader(body)
	request, err := http.NewRequest("POST", postPushDataUrl, reader)
	if err != nil {
		return
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return
	}
	defer resp.Body.Close()
}
