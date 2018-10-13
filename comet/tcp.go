package main

import (
	"bufio"
	"im/libs/define"
	"im/libs/proto"
	itime "im/libs/time"
	"net"
	"time"
)

func InitTCPServ(addrs []string, accept int) (err error) {
	var (
		bind     string
		listener *net.TCPListener
		addr     *net.TCPAddr
	)
	for _, bind = range addrs {
		if addr, err = net.ResolveTCPAddr("tcp", bind); err != nil {
			return
		}
		if listener, err = net.ListenTCP("tcp", addr); err != nil {
			return
		}
		for i := 0; i < accept; i++ {
			go acceptTCP(DefaultServer, listener)
		}
	}
	return
}

func acceptTCP(server *Server, lis *net.TCPListener) {
	var (
		conn *net.TCPConn
		err  error
		r    int
	)
	for {
		if conn, err = lis.AcceptTCP(); err != nil {
			return
		}
		if err = conn.SetReadBuffer(server.Options.TCPRcvbuf); err != nil {
			return
		}
		if err = conn.SetWriteBuffer(server.Options.TCPSendBuf); err != nil {
			return
		}
		go serveTCP(server, conn, r)
		if r++; r == maxInt {
			r = 0
		}
	}
}

func serveTCP(server *Server, conn *net.TCPConn, r int) {
	var (
		tr = server.round.Timer(r)
		//lAddr = conn.LocalAddr().String()
		rAddr = conn.RemoteAddr().String()
	)
	server.serveTCP(conn, tr, rAddr)
}

func (server *Server) serveTCP(conn *net.TCPConn, tr *itime.Timer, rAddr string) {
	var (
		err error
		key string
		hb  time.Duration
		p   *proto.Proto
		b   *Bucket
		trd *itime.TimerData
		ch  = NewChannel(server.Options.CliProto, server.Options.SvrProto, define.NoRoom)
	)
	ch.Reader = bufio.NewReader(conn)
	ch.Writer = bufio.NewWriter(conn)
	var (
		rr = ch.Reader
		wr = ch.Writer
	)
	trd = tr.Add(server.Options.HandshakeTimeout, func() {
		conn.Close()
	})

	if p, err = ch.CliProto.Set(); err == nil {
		if key, ch.RoomID, hb, err = server.authTCP(rr, wr, p, rAddr); err == nil {
			b = server.Bucket(ch.RoomID)
			err = b.Put(key, ch)
		}
	}
	if err != nil {
		conn.Close()
		tr.Del(trd)
		return
	}

	trd.Key = key
	tr.Set(trd, hb)

	go server.dispatchTCP(key, conn, wr, ch)

	for {
		if p, err = ch.CliProto.Set(); err != nil {
			break
		}
		ch.CliProto.SetAdv()
		if err = p.ReadTCP(rr); err != nil {
			break
		}
		tr.Set(trd, hb)
		if p.Operation == define.OP_HEARTBEAT {
			//心跳包无需回复
			//p.Body = nil
			//p.Operation = define.OP_HEARTBEAT_REPLY
			//ch.Signal()
		} else {
			if err = server.operator.Send(key, p); err != nil {
				break
			}
		}
		ch.CliProto.GetAdv()
	}
	conn.Close()
	b.Del(key)
	tr.Del(trd)
	ch.Close()
	server.operator.Disconnect(key)
}

func (server *Server) dispatchTCP(key string, conn *net.TCPConn, wr *bufio.Writer, ch *Channel) {
	var (
		err    error
		finish bool
	)
	for {
		var p = ch.Ready()
		switch p {
		case proto.ProtoFinish:
			finish = true
			goto failed
		default:
			if err = p.WriteTCP(wr); err != nil {
				goto failed
			}
		}
		if err = wr.Flush(); err != nil {
			break
		}
	}
failed:
	conn.Close()
	for !finish {
		finish = (ch.Ready() == proto.ProtoFinish)
	}
}

func (server *Server) authTCP(rr *bufio.Reader, wr *bufio.Writer, p *proto.Proto, rAddr string) (key string, rid int32, heartbeat time.Duration, err error) {

	if err = p.ReadTCP(rr); err != nil {
		return
	}

	if p.Operation != define.OP_AUTH {
		err = ErrOperation
		return
	}

	if key, rid, heartbeat, err = server.operator.Connect(p, rAddr); err != nil {
		return
	}
	p.Body = nil
	p.Operation = define.OP_AUTH_REPLY
	if err = p.WriteTCP(wr); err != nil {
		return
	}
	err = wr.Flush()
	return
}
