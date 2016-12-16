package client

import (
	"strconv"
	"time"

	"github.com/qxnw/grpc4ars/pb"
	"github.com/qxnw/lib4go/logger"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

//Client client
type Client struct {
	address       string
	conn          *grpc.ClientConn
	opts          *clientOption
	client        pb.ARSClient
	longTicker    *time.Ticker
	lastRequest   time.Time
	hasRunChecker bool
	IsConnect     bool
	isClose       bool
}

type clientOption struct {
	heartbeat         bool
	heartbeatTicker   time.Duration
	connectionTimeout time.Duration
	log               logger.ILogger
}

//ClientOption 客户端配置选项
type ClientOption func(*clientOption)

//WithHeartbeat 定时发送心跳数据
func WithHeartbeat() ClientOption {
	return func(o *clientOption) {
		o.heartbeat = true
	}
}

//WithCheckTicker 连接状态检查间隔时间
func WithCheckTicker(t time.Duration) ClientOption {
	return func(o *clientOption) {
		o.heartbeatTicker = t
	}
}

//WithConnectionTimeout 网络连接超时时长
func WithConnectionTimeout(t time.Duration) ClientOption {
	return func(o *clientOption) {
		o.connectionTimeout = t
	}
}

//WithLogger 设置日志记录器
func WithLogger(log logger.ILogger) ClientOption {
	return func(o *clientOption) {
		o.log = log
	}
}

//NewClient 创建客户端
func NewClient(address string, opts ...ClientOption) *Client {
	client := &Client{address: address, opts: &clientOption{heartbeatTicker: time.Second * 15, connectionTimeout: time.Second * 3}}
	for _, opt := range opts {
		opt(client.opts)
	}
	return client
}

//Connect 连接服务器，如果当前无法连接系统会定时自动重连
func (c *Client) Connect() (b bool) {
	if c.IsConnect {
		return
	}
	var err error
	c.conn, err = grpc.Dial(c.address, grpc.WithInsecure(), grpc.WithTimeout(c.opts.connectionTimeout))
	if err != nil {
		c.IsConnect = false
		return c.IsConnect
	}
	c.client = pb.NewARSClient(c.conn)
	if c.opts.heartbeat && !c.hasRunChecker {
		c.hasRunChecker = true
		go c.connectCheck()
	}
	//检查是否已连接到服务器
	response, er := c.client.Heartbeat(context.Background(), &pb.HBRequest{Ping: 0})
	c.IsConnect = er == nil && response.Pong == 0
	return c.IsConnect
}

//Request 发送请求
func (c *Client) Request(session string, service string, data string) (status int, result string, err error) {
	c.lastRequest = time.Now()
	response, err := c.client.Request(context.Background(), &pb.RequestContext{Session: session, Sevice: service, Input: data})
	if err != nil {
		c.IsConnect = false
		return
	}
	status = int(response.Status)
	result = response.GetResult()
	c.IsConnect = true
	return
}

//connectCheck 网络连接状态检查
func (c *Client) connectCheck() {
	c.longTicker = time.NewTicker(c.opts.heartbeatTicker)
	for {
		select {
		case <-c.longTicker.C:
			oping, _ := strconv.Atoi(time.Now().Format("150405"))
			ping := int32(oping)
			response, err := c.client.Heartbeat(context.Background(), &pb.HBRequest{Ping: ping})
			c.IsConnect = err == nil && response.Pong == ping
			c.logInfof("[心跳]%s %d %v", c.address, oping, c.IsConnect)
			c.lastRequest = time.Now()
		}
	}
}

//logInfof 日志记录
func (c *Client) logInfof(format string, msg ...interface{}) {
	if c.opts.log == nil {
		return
	}
	c.opts.log.Infof(format, msg...)
}

//Close 关闭连接
func (c *Client) Close() {
	c.isClose = true
	if c.longTicker != nil {
		c.longTicker.Stop()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
