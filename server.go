package main

import (
	"fmt"
	"io"
	"net"
	"sync"
)

// 创建一个server对象
type Server struct {
	Ip   string
	Port int

	// 在线用户的列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	// 消息广播的channle
	Message chan string
}

// 创建一个Server 接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}

	return server
}

func (s *Server) Handler(conn net.Conn) {
	fmt.Println("链接建立成功~~~")

	user := NewUser(conn,s)

	user.Online()

	// 接收客户端发送的消息
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			// 若读取的字节数为0，默认值
			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}
			// 提取用户的消息（去除'\n'）
			msg:=string(buf[:n-1])

			// 用户针对msg进行消息处理
			user.DoMessage(msg)
		}
	}()

	// 当前handler阻塞
	select {}

}

// 监听Message广播消息channle的goroutine，一旦有消息就发送给全部的在线User
func (s *Server) ListenMessage() {
	for {
		msg := <-s.Message

		// 将msg发送给全部在在线的User
		s.mapLock.Lock()
		for _, m := range s.OnlineMap {
			m.C <- msg
		}
		s.mapLock.Unlock()
	}
}

// 广播消息的方法
func (s *Server) Broadcast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	s.Message <- sendMsg
}

// 启动服务器接口
func (s *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.Ip, s.Port))
	if err != nil {
		fmt.Println("net.Listen err: ", err)
		return
	}
	// 关闭监听程序
	defer listener.Close()

	// 启动监听广播的协程
	go s.ListenMessage()

	for {
		// accept
		conn, err2 := listener.Accept()
		if err2 != nil {
			fmt.Println("listener accept err: ", err2)
			continue
		}

		go s.Handler(conn)
	}
}
