package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn

	flag int //当前Client所处于的模式
}

func NewClient(serverIp string, serverPort int) *Client {
	// 创建客户端
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999, //新建客户端默认999
	}
	// 链接server
	c, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial err", err)
		return nil
	}
	client.conn = c

	// 返回对象
	return client
}

var serverIp string
var serverPort int

// ./client -ip 127.0.0.1 -port 8888
func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器IP地址(默认是127.0.0.1)")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器Port地址(默认是8888)")
}

func main() {
	// 命令行解析
	flag.Parse()

	client := NewClient("127.0.0.1", 8888)
	if client == nil {
		fmt.Println(">>>>> 链接服务器失败...")
		return
	}
	fmt.Println(">>>>> 链接服务器成功...")

	//单独开启一个goroutine去处理server的回执消息
	go client.ShowResponse()
	// 启动客户端业务
	client.Run()
}

// 菜单
func (client *Client) menu() bool {
	var flag int
	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")

	fmt.Scanln(&flag)
	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println(">>>>>>请输入合法范围的数字>>>>>>")
		return false
	}
}

// 处理server回应的消息，直接显示到标准输出
func (client *Client) ShowResponse() {
	//一旦client.conn有数据，就直接copy到stdout标准输出上, 永久阻塞监听
	io.Copy(os.Stdout, client.conn)
}

func (client *Client) Run() {
	// 如果flag不为0，则一直循环
	for client.flag != 0 {
		// 输入数字不合法，循环
		for client.menu() != true {
		}

		// 根据不同模式进入对应的业务
		switch client.flag {
		case 1:
			// fmt.Println("公聊业务处理。。。")
			client.PublicChat()
			break
		case 2:
			// fmt.Println("私聊业务处理。。。")
			client.PrivateChat()
			break
		case 3:
			// fmt.Println("更改用户名业务处理。。。")
			client.UpdateName()
			break
		default:
			fmt.Println("-----程序退出----")
		}
	}
}

// -----更改用户名-----
func (client *Client) UpdateName() bool {
	fmt.Print(">>>请输入用户名：")
	fmt.Scanln(&client.Name)
	sendMsg := "rename|" + client.Name + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn write err:", err)
		return false
	}
	return true
}

// -------私聊模式-------
func (client *Client) SelectUsers() {
	sendMsg := "who\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn write err:", err)
		return
	}
}
func (client *Client) PrivateChat() {
	var remoteName string
	var chatMsg string

	client.SelectUsers()
	fmt.Println(">>>>>>请输入聊天对象【用户名】,exit退出")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		fmt.Println(">>>>>>请输入聊天信息,exit退出")
		fmt.Scanln(&chatMsg)
		for chatMsg != "exit" {
			// 聊天信息不能为空
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n"
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn write err:", err)
					break
				}
			}
			chatMsg = ""
			fmt.Println(">>>>>>请输入聊天信息,exit退出")
			fmt.Scanln(&chatMsg)
		}
		client.SelectUsers()
		fmt.Println(">>>>>>请输入聊天对象【用户名】,exit退出")
		fmt.Scanln(&remoteName)

	}
}

// -----公聊模式------
func (client *Client) PublicChat() {
	var chatMsg string
	fmt.Println(">>>> 请输入聊天内容，输入exit退出")
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		// 若消息不为空则发送给服务器
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn write err:", err)
				break
			}
		}
		chatMsg = ""
		fmt.Println(">>>> 请输入聊天内容，输入exit退出")
		fmt.Scanln(&chatMsg)
	}
}
