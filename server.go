package main

import (
	"fmt"
	"io"
	"net"
	"runtime"
	"sync"
	"time"
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

	user := NewUser(conn, s)

	user.Online()

	//监听用户是否活跃的channel
	isLive := make(chan bool)

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
			msg := string(buf[:n-1])

			// 用户针对msg进行消息处理
			user.DoMessage(msg)

			// 用户的任意消息，都代表当前用户时活跃的
			isLive <- true
		}
	}()

	// 查看OnlineMap
	// fmt.Printf("%+v\n", s.OnlineMap)

	// 当前handler阻塞
	for {
		select {
		case <-isLive:
		case <-time.After(time.Minute * 30):
			// 已经超时，将当前的User强制关闭
			user.SendMsg("你被踢了")
			// 销毁channle资源
			close(user.C)
			// 关闭连接
			conn.Close()
			// 退出当前handler
			// return
			runtime.Goexit()
		}
	}

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
