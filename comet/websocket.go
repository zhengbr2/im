package main

import (
	"crypto/tls"
	"im/libs/define"
	"im/libs/proto"
	itime "im/libs/time"
	"math/rand"

	"net"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func InitWebsocketServ(addrs []string) (err error) {
	var (
		bind         string
		listener     *net.TCPListener
		addr         *net.TCPAddr
		httpServeMux = http.NewServeMux()
		server       *http.Server
	)
	httpServeMux.HandleFunc("/ws", ServeWebSocket)

	for _, bind = range addrs {
		if addr, err = net.ResolveTCPAddr("tcp", bind); err != nil {
			return
		}

		if listener, err = net.ListenTCP("tcp", addr); err != nil {
			return
		}
		server = &http.Server{Handler: httpServeMux}

		go func(host string) {
			if err = server.Serve(listener); err != nil {
				panic(err)
			}
		}(bind)
	}
	return
}

func InitWebsocketWithTLS(addrs []string, cert, priv string) (err error) {
	var (
		httpServeMux = http.NewServeMux()
	)
	httpServeMux.HandleFunc("/wss", ServeWebSocket)
	config := &tls.Config{}
	config.Certificates = make([]tls.Certificate, 1)
	if config.Certificates[0], err = tls.LoadX509KeyPair(cert, priv); err != nil {
		return
	}
	for _, bind := range addrs {
		server := &http.Server{Addr: bind, Handler: httpServeMux}
		server.SetKeepAlivesEnabled(true)
		go func(host string) {
			ln, err := net.Listen("tcp", host)
			if err != nil {
				return
			}

			tlsListener := tls.NewListener(ln, config)
			if err = server.Serve(tlsListener); err != nil {
				return
			}
		}(bind)
	}
	return
}

func ServeWebSocket(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	ws, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		return
	}
	defer ws.Close()
	var (
		tr = DefaultServer.round.Timer(rand.Int())
		//lAddr = ws.LocalAddr().String()
		rAddr = ws.RemoteAddr().String()
	)
	DefaultServer.serveWebsocket(ws, tr, rAddr)
}

func (server *Server) serveWebsocket(conn *websocket.Conn, tr *itime.Timer, rAddr string) {
	var (
		err error
		key string
		hb  time.Duration
		p   *proto.Proto
		b   *Bucket
		trd *itime.TimerData
		ch  = NewChannel(server.Options.CliProto, server.Options.SvrProto, define.NoRoom)
	)
	trd = tr.Add(server.Options.HandshakeTimeout, func() {
		conn.Close()
	})

	if p, err = ch.CliProto.Set(); err == nil {
		if key, ch.RoomID, hb, err = server.authTCPWebsocket(conn, p, rAddr); err == nil {
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

	go server.dispatchWebsocket(key, conn, ch)

	for {
		if p, err = ch.CliProto.Set(); err != nil {
			break
		}
		ch.CliProto.SetAdv()
		if err = p.ReadWebsocket(conn); err != nil {
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

func (server *Server) dispatchWebsocket(key string, conn *websocket.Conn, ch *Channel) {
	var (
		p   *proto.Proto
		err error
	)
	for {
		p = ch.Ready()
		switch p {
		case proto.ProtoFinish:
			goto failed
		default:
			if err = p.WriteWebsocket(conn); err != nil {
				goto failed
			}
		}
	}
failed:
	conn.Close()
	for {
		if p == proto.ProtoFinish {
			break
		}
		p = ch.Ready()
	}
}

func (server *Server) authTCPWebsocket(conn *websocket.Conn, p *proto.Proto, rAddr string) (key string, rid int32, heartbeat time.Duration, err error) {
	if err = p.ReadWebsocket(conn); err != nil {
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
	err = p.WriteWebsocket(conn)
	return
}
