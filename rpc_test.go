package grpc4ars

import (
	"testing"

	"time"

	"fmt"

	"github.com/qxnw/grpc4ars/client"
	"github.com/qxnw/grpc4ars/server"
)

// TestNew 测试初始化一个pool
func TestNew(t *testing.T) {
	svr := server.NewServer(func(session string, svs string, data string) (status int, result string, err error) {
		status = 100
		result = svs
		return
	})
	go func() {
		if err := svr.Start(":10160"); err != nil {
			t.Error(err)
		}
	}()

	client := client.NewClient()
	if e := client.ConnectTimeout(":10162", time.Second*3); e != nil {

		t.Error(e)
		return
	}
	fmt.Println("s:", time.Now())
	s, result, err := client.Request("123455666", "svname", "{}")
	//	fmt.Println(svname)
	if err != nil {
		t.Error(err)
	}
	if s != 100 || result != "svname" {
		t.Error("数据有误")
	}
	fmt.Println("e:", time.Now())
	/*
		mu := sync.WaitGroup{}
		for i := 0; i < 10000; i++ {
			mu.Add(1)
			go func(i int) {
				svname := fmt.Sprintf("svs:%d", i)
				s, result, err := client.Request("123455666", svname, "{}")
				//	fmt.Println(svname)
				if err != nil {
					t.Error(err)
				}
				if s != 100 || result != svname {
					t.Error("数据有误")
				}
				mu.Done()

			}(i)
		}
		mu.Wait()
	*/
}
