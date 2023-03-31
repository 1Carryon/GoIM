package user

import "net"

type User struct {
	Name string
	Addr string
	C    chan string //当前用户绑定的channel
	conn net.Conn
}

// 创建一个用户的api
func NewUser(conn net.Conn) *User {
	userAddr := conn.RemoteAddr().String() //拿到当前的地址链接

	user := &User{
		Name: "user_" + userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,
	}

	// 启动监听当前user channel消息的协程
	go user.ListenMessage()

	return user
}

// 监听当前User channel的方法，一旦有消息，就直接发送给对应客户端
func (u *User) ListenMessage() {
	for {
		msg := <-u.C
		u.conn.Write([]byte(msg + "\n"))
	}
}
