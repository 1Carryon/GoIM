package main

import (
	"net"
	"strings"
)

type User struct {
	Name string
	Addr string
	C    chan string //当前用户绑定的channel
	conn net.Conn

	// user类型新增sever关联--->用于用户业务层封装
	server *Server
}

// 创建一个用户的api
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String() //拿到当前的地址链接

	user := &User{
		Name: "user_" + userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,

		server: server,
	}

	// 启动监听当前user channel消息的协程
	go user.ListenMessage()

	return user
}

// 用户上线的业务
func (u *User) Online() {
	// 用户上线,将用户加入onlineMap中
	u.server.mapLock.Lock()
	u.server.OnlineMap[u.Name] = u
	u.server.mapLock.Unlock()

	// 广播当前用户上线消息
	u.server.Broadcast(u, "已上线")
}

// 用户下线的业务
func (u *User) Offline() {
	// 用户下线，将用户从onlineMap中删除
	u.server.mapLock.Lock()
	// 调用delete函数将user从map中删除
	delete(u.server.OnlineMap, u.Name)
	u.server.mapLock.Unlock()

	// 广播当前用户下线消息
	u.server.Broadcast(u, "下线")
}

// 给当前的User对应的客户端发送消息
func (u *User) SendMsg(msg string) {
	u.conn.Write([]byte(msg))

}

// 用户处理消息的业务
func (u *User) DoMessage(msg string) {
	if msg == "who" { //对“who”指令处理
		// 查询当前在线的用户都有哪些
		u.server.mapLock.Lock()
		for _, user := range u.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "在线...\n"
			u.SendMsg(onlineMsg)
		}
		u.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		//消息格式 rename|张三
		newName := strings.Split(msg, "|")[1]

		// 判断name是否存在
		_, ok := u.server.OnlineMap[newName]
		if ok {
			u.SendMsg("当前用户名已存在\n")
		} else {
			// 更改map中的指针引向
			u.server.mapLock.Lock()
			delete(u.server.OnlineMap, u.Name)
			u.server.OnlineMap[newName] = u
			u.server.mapLock.Unlock()

			// 实际更改
			u.Name = newName
			u.SendMsg("您已更新用户名：" + u.Name + "\n")
		}
	} else {
		u.server.Broadcast(u, msg)
	}

}

// 监听当前User channel的方法，一旦有消息，就直接发送给对应客户端
func (u *User) ListenMessage() {
	for {
		msg := <-u.C
		u.conn.Write([]byte(msg + "\n"))
	}
}
