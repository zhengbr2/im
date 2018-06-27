package main

import (
	"time"

	"github.com/globalsign/mgo"
	//"github.com/globalsign/mgo/bson"
)

var (
	session *mgo.Session
)

func init() {
	dialInfo := &mgo.DialInfo{
		Addrs:    []string{"localhost"},
		Direct:   false,
		Timeout:  time.Second * 5,
		Database: "gkim",
		//Source:   "admin",
		//Username: "",
		//Password: "",
		//PoolLimit: 4096, // Session.SetPoolLimit
	}
	var (
		err error
	)
	session, err = mgo.DialWithInfo(dialInfo)
	if nil != err {
		panic(err)
	}
	//defer session.Close()
	session.SetMode(mgo.Monotonic, true)
}

type DB struct {
	session *mgo.Session
}

func (d *DB) C(c string) *mgo.Collection {
	if d.session == nil {
		d.session = session.Copy()
	}
	return d.session.DB("gkim").C(c)
}

func (d *DB) Close() {
	if d.session != nil {
		d.session.Close()
		d.session = nil
	}
}
