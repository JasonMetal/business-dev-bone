package server

import (
	"business-dev-bone/pkg/component-base/log"
	"github.com/xtaci/kcp-go/v5"
)

type KCPHandler interface {
	ServeKCP(conn *kcp.UDPSession)
}

type KCPServer struct {
	addr    string
	handler KCPHandler
}

func NewKCPServer(addr string, handler KCPHandler) *KCPServer {
	return &KCPServer{addr: addr, handler: handler}
}

func (s *KCPServer) Run() error {
	go func() {
		log.Infof("Start to listening the incoming requests on KCP address: %s", s.addr)
		listener, err := kcp.ListenWithOptions(s.addr, nil, 0, 0)
		if err != nil {
			log.Fatal(err.Error())
		}

		for {
			conn, err := listener.AcceptKCP()
			if err != nil {
				log.Errorf("accept kcp connection failed: %v", err)
				continue
			}

			log.Infof("new kcp connection: %d, remote: %v, local: %v", conn.GetConv(), conn.RemoteAddr().String(), conn.LocalAddr().String())

			// 配置KCP参数
			conn.SetNoDelay(1, 10, 2, 1)       // 快速模式
			conn.SetStreamMode(true)           // if you use UDPSession, UDPSession works only in stream mode, but it does make a difference if you use kcp.go directly.
			conn.SetWriteDelay(false)          // 禁用写延迟，立即发送数据
			conn.SetACKNoDelay(true)           // 启用无延迟ACK，提高实时性
			conn.SetWindowSize(1024*4, 1024*4) // 参考unity sdk设置
			conn.SetMtu(1200)                  // 参考unity sdk设置
			go s.handler.ServeKCP(conn)
		}
	}()

	return nil
}
