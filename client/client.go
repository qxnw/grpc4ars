package client

import (
	"net"
	"time"

	"github.com/qxnw/grpc4ars/pb"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

//Client client
type Client struct {
	conn   *grpc.ClientConn
	client pb.ARSClient
}

//NewClient 创建客户端
func NewClient() *Client {
	return &Client{}
}

//ConnectTimeout 连接服务器
func (c *Client) ConnectTimeout(address string, timeout time.Duration) (err error) {
	c.conn, err = grpc.Dial(address, grpc.WithInsecure(), grpc.WithTimeout(timeout), grpc.WithDialer(
		func(address string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("tcp", address, timeout)
		}))
	if err != nil {
		return
	}
	c.client = pb.NewARSClient(c.conn)
	return
}

//Connect 连接服务器
func (c *Client) Connect(address string) (err error) {
	c.conn, err = grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return
	}
	c.client = pb.NewARSClient(c.conn)
	return
}

//Request 发送请求
func (c *Client) Request(session string, service string, data string) (status int, result string, err error) {
	response, err := c.client.Request(context.Background(), &pb.RequestContext{Session: session, Sevice: service, Input: data})
	if err != nil {
		return
	}
	status = int(response.Status)
	result = response.GetResult()
	return
}

//Close 关闭连接
func (c *Client) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}
