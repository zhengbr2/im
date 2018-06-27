package xrpc

import (
	"errors"
	"im/libs/proto"
	"net"
	"net/rpc"
	"time"
)

const (
	dialTimeout  = 5 * time.Second
	callTimeout  = 3 * time.Second
	pingDuration = 1 * time.Second
)

var (
	ErrRpc        = errors.New("rpc is not available")
	ErrRpcTimeout = errors.New("rpc call timeout")
)

type ClientOptions struct {
	Proto string
	Addr  string
}

type Client struct {
	*rpc.Client
	options ClientOptions
	quit    chan struct{}
	err     error
}

func Dial(options ClientOptions) (c *Client) {
	c = new(Client)
	c.options = options
	c.dial()
	return
}

func (c *Client) dial() (err error) {
	var conn net.Conn
	conn, err = net.DialTimeout(c.options.Proto, c.options.Addr, dialTimeout)
	if err == nil {
		c.Client = rpc.NewClient(conn)
	}
	return
}

func (c *Client) Call(serviceMethod string, args interface{}, reply interface{}) (err error) {
	if c.Client == nil {
		err = ErrRpc
		return
	}
	select {
	case call := <-c.Client.Go(serviceMethod, args, reply, make(chan *rpc.Call, 1)).Done:
		err = call.Error
	case <-time.After(callTimeout):
		err = ErrRpcTimeout
	}
	return
}

func (c *Client) Error() error {
	return c.err
}

func (c *Client) Close() {
	c.quit <- struct{}{}
}

func (c *Client) Ping(serviceMethod string) {
	var (
		arg   = proto.NoArg{}
		reply = proto.NoReply{}
		err   error
	)
	for {
		select {
		case <-c.quit:
			goto closed
			return
		default:
		}
		if c.Client != nil && c.err == nil {
			if err = c.Call(serviceMethod, &arg, &reply); err != nil {
				c.err = err
				if err != rpc.ErrShutdown {
					c.Client.Close()
				}
			}
		} else {
			if err = c.dial(); err == nil {
				c.err = nil
			}
		}
		time.Sleep(pingDuration)
	}
closed:
	if c.Client != nil {
		c.Client.Close()
	}
}
