package client

import (
	"fmt"
	"sync/atomic"

	"errors"

	"github.com/qxnw/lib4go/concurrent/cmap"
)

//ClientPool 客户端连接池
type ClientPool struct {
	address string
	opt     *poolOption
	clients chan *Client
	cache   cmap.ConcurrentMap
}

type cacheClient struct {
	Concurrent int32
	Key        string
}
type poolOption struct {
	concurrent int
	minConn    int
}

//PoolOption 连接池配置选项
type PoolOption func(*poolOption)

//WithMaxConcurrent 每个连接最大并发数
func WithMaxConcurrent(concurrent int) PoolOption {
	return func(o *poolOption) {
		o.concurrent = concurrent
	}
}

//WithMinConnection 最小客户端连接数
func WithMinConnection(min int) PoolOption {
	return func(o *poolOption) {
		o.minConn = min
	}
}

//NewClientPool 初始化客户端连接池
func NewClientPool(address string, opts ...PoolOption) (p *ClientPool) {
	p = &ClientPool{opt: &poolOption{minConn: 1, concurrent: 10000}, address: address}
	p.cache = cmap.New()
	for _, opt := range opts {
		opt(p.opt)
	}
	p.clients = make(chan *Client, p.opt.minConn)
	for i := 0; i < p.opt.minConn; i++ {
		client, _ := p.create()
		p.clients <- client
	}
	return
}

//get 从缓存池中获取一个可用的连接，并立即放回连接池，如果没有可用的则立即创建一个连接
func (c *ClientPool) get() (client *Client, cache *cacheClient, needBack bool, err error) {
	hasGet := make(map[string]string)
START:
	select {
	//从缓存中获取一个连接并立即放回
	case client, ok := <-c.clients:
		if !ok {
			err = errors.New("连接池已关闭")
			return nil, nil, false, err
		}
		key := fmt.Sprintf("%p", client)
		cacheC, _ := c.cache.Get(key)
		cache = cacheC.(*cacheClient)
		select {
		case c.clients <- client:
			needBack = false
		default:
			needBack = true
		}
	default:
		//连接池中没有可用连接则立即创建一个新的连接
		needBack = true
		client, cache = c.create()
	}

	//检查连接是否可用
	//如果并发数是否大于配置数，如果已大于配置数，检查该连接是否已获取过则创建新连接，否则重新从缓存中获取
	if int(cache.Concurrent) >= c.opt.concurrent {
		if _, ok := hasGet[cache.Key]; ok {
			if needBack {
				c.back(client)
			}
			client, cache = c.create()
			needBack = true
			return
		}
		goto START
	}
	return
}
func (c *ClientPool) back(client *Client) {
	select {
	case c.clients <- client:
	default:
		client.Close()
	}
}
func (c *ClientPool) create() (*Client, *cacheClient) {
	client := NewClient(c.address)
	client.Connect()
	key := fmt.Sprintf("%p", client)
	_, cacheC, _ := c.cache.SetIfAbsentCb(key, func(input ...interface{}) (interface{}, error) {
		return &cacheClient{Concurrent: 0, Key: key}, nil
	})
	return client, cacheC.(*cacheClient)
}

//Request 发送服务器请求
func (c *ClientPool) Request(session string, service string, data string) (status int, result string, err error) {
	//1. 获取一个可用的连接
	client, cache, back, err := c.get()
	if err != nil {
		return
	}
	atomic.AddInt32(&cache.Concurrent, 1)
	//2. 发送请求,并将引用引数减1
	status, result, err = client.Request(session, service, data)
	atomic.AddInt32(&cache.Concurrent, -1)
	if back {
		c.back(client)
	}
	return
}

//Close 半闭当前连接池
func (c *ClientPool) Close() {
	close(c.clients)
	for {
		select {
		case client := <-c.clients:
			client.Close()
		default:
			return
		}
	}
}
