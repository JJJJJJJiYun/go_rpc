package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/JJJJJJJiYun/go_rpc/codec"
	"github.com/JJJJJJJiYun/go_rpc/opt"
)

type Call struct {
	Seq           uint64
	ServiceMethod string
	Args          interface{}
	Reply         interface{}
	Error         error
	Done          chan *Call
}

func (c *Call) done() {
	c.Done <- c
}

type Client struct {
	cc       codec.Codec
	opt      *opt.Option
	sending  sync.Mutex
	header   codec.Header
	mu       sync.Mutex
	seq      uint64
	pending  map[uint64]*Call
	closing  bool
	shutdown bool
}

var ErrShutdown = fmt.Errorf("connection is shut down")

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closing {
		return ErrShutdown
	}
	c.closing = true
	return c.cc.Close()
}

func (c *Client) IsAvailable() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return !c.shutdown && !c.closing
}

func (c *Client) registerCall(call *Call) (uint64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closing || c.shutdown {
		return 0, ErrShutdown
	}
	call.Seq = c.seq
	c.pending[call.Seq] = call
	c.seq++
	return call.Seq, nil
}

func (c *Client) removeCall(seq uint64) (call *Call) {
	c.mu.Lock()
	defer c.mu.Unlock()
	call = c.pending[seq]
	delete(c.pending, seq)
	return
}

func (c *Client) terminateCall(err error) {
	c.sending.Lock()
	defer c.sending.Unlock()
	c.mu.Lock()
	defer c.mu.Unlock()
	c.shutdown = true
	for _, c := range c.pending {
		c.Error = err
		c.done()
	}
}

func (c *Client) receive() {
	var err error
	for err == nil {
		var h codec.Header
		if err = c.cc.ReadHeader(&h); err != nil {
			break
		}
		call := c.removeCall(h.Seq)
		switch {
		case call == nil:
			err = c.cc.ReadBody(nil)
		case h.Error != "":
			call.Error = fmt.Errorf(h.Error)
			err = c.cc.ReadBody(nil)
			call.done()
		default:
			err = c.cc.ReadBody(call.Reply)
			if err != nil {
				call.Error = fmt.Errorf("reading body err: %v", err)
			}
			call.done()
		}
	}
	c.terminateCall(err)
}

func (c *Client) send(call *Call) {
	c.sending.Lock()
	defer c.sending.Unlock()
	seq, err := c.registerCall(call)
	if err != nil {
		call.Error = err
		call.done()
		return
	}
	c.header.ServiceMethod = call.ServiceMethod
	c.header.Seq = seq
	if err := c.cc.Write(&c.header, call.Args); err != nil {
		call := c.removeCall(seq)
		if call != nil {
			call.Error = err
			call.done()
		}
	}
}

func (c *Client) Go(serviceMethod string, args, reply interface{}, done chan *Call) *Call {
	if done == nil {
		// 异步的话默认十个
		done = make(chan *Call, 10)
	} else if cap(done) == 0 {
		log.Panic("rpc client: done channel is unbuffered")
	}
	call := &Call{
		ServiceMethod: serviceMethod,
		Args:          args,
		Reply:         reply,
		Done:          done,
	}
	c.send(call)
	return call
}

func (c *Client) Call(serviceMethod string, args, reply interface{}) error {
	call := <-c.Go(serviceMethod, args, reply, make(chan *Call, 1)).Done
	return call.Error
}

func NewClient(conn net.Conn, option *opt.Option) (*Client, error) {
	f := codec.NewCodecFuncMap[option.CodecType]
	if f == nil {
		err := fmt.Errorf("invalid codec type: %s", option.CodecType)
		log.Println("rpc client: codec error: ", err)
		return nil, err
	}
	if err := json.NewEncoder(conn).Encode(option); err != nil {
		log.Println("rpc client: options error: ", err)
		_ = conn.Close()
		return nil, err
	}
	return newClientCodec(f(conn), option), nil
}

func newClientCodec(cc codec.Codec, option *opt.Option) *Client {
	client := &Client{
		cc:      cc,
		opt:     option,
		seq:     1,
		pending: make(map[uint64]*Call),
	}
	go client.receive()
	return client
}

func parseOptions(opts ...*opt.Option) (*opt.Option, error) {
	if len(opts) == 0 || opts[0] == nil {
		return opt.DefaultOption, nil
	}
	if len(opts) != 1 {
		return nil, fmt.Errorf("number of options is more than 1")
	}
	option := opts[0]
	option.MagicNumber = opt.DefaultOption.MagicNumber
	if option.CodecType == "" {
		option.CodecType = opt.DefaultOption.CodecType
	}
	return option, nil
}

func Dial(network, address string, opts ...*opt.Option) (client *Client, err error) {
	option, err := parseOptions(opts...)
	if err != nil {
		return nil, err
	}
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	defer func() {
		if client == nil {
			_ = conn.Close()
		}
	}()
	return NewClient(conn, option)
}
