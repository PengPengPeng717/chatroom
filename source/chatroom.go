package main

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

type User struct {
	name string
	id   string
	msg  chan string
	done chan bool
}

// 创建一个全局的map,用于保存所有用户
var (
	allusers      = make(map[string]User, 10)
	allusersMutex = sync.RWMutex{}
)

// 定义一个message全局通道，用于接收任何人发送的消息
var message = make(chan string, 10)

func main() {
	// 创建服务器
	listener, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		fmt.Println("listen err:", err)
		return
	}
	// defer listener.Close()
	// 启动全局唯一的go程，负责监听message通道，并把消息广播给所有客户端
	// go broadcast()
	fmt.Println("服务器启动成功")
	go broadcast()
	for {

		//监听
		fmt.Println("监听中")
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Accept err:", err)
			return
		}
		// defer conn.Close()

		//建立连接
		fmt.Println("客户端已连接：", conn.RemoteAddr().String())

		//启动处理业务的go程
		go handler(conn)
	}

}

// 处理具体业务
func handler(conn net.Conn) {
	defer conn.Close()

	fmt.Println("处理业务")
	//TODO
	// 创建一个缓冲区
	clientArr := conn.RemoteAddr().String()
	fmt.Println("客户端已连接：", clientArr)
	newUser := User{
		name: clientArr,
		id:   clientArr,
		msg:  make(chan string),
		done: make(chan bool),
	}

	// 添加用户到map
	allusersMutex.Lock()
	allusers[newUser.id] = newUser
	allusersMutex.Unlock()

	//定义一个退出信号，用于监听client退出
	var isQuit = make(chan bool)
	go watchExit(&newUser, conn, isQuit)

	//启动go程，负责将msg信息返回给客户端
	go writeBackToClient(&newUser, conn)

	//向message写入数据，当前用户上线的消息，用于通知所有人
	loginInfo := fmt.Sprintf("用户[%s]:[%s]上线了\n", newUser.id, newUser.name)
	message <- loginInfo

	defer func() {
		// 确保用户在退出时从全局映射中删除
		allusersMutex.Lock()
		if _, exists := allusers[newUser.id]; exists {
			delete(allusers, newUser.id)
			// 向所有人发送下线通知
			logoutInfo := fmt.Sprintf("用户[%s]:[%s]下线了\n", newUser.id, newUser.name)
			message <- logoutInfo
		}
		allusersMutex.Unlock()
	}()

	for {
		buf := make([]byte, 1024)
		// 循环读取客户端发送的数据
		// 读取数据
		n, err := conn.Read(buf)
		if n == 0 {
			fmt.Println("用户已下线")
			// 发送退出信号
			isQuit <- true
			return
		}
		if err != nil {
			fmt.Println("read err:", err)
			return
		}

		// 处理接收到的数据并广播给所有用户
		fmt.Println("接收到的数据：", string(buf[:n-1]), "长度：", n)

		// 检查是否应该退出
		select {
		case <-newUser.done:
			fmt.Println("收到退出信号，停止处理消息")
			return
		default:
		}

		// --················-- 业务逻辑 ---······················-
		//1.查询当前所有的用户who
		// a.先判读接收的数据是不是who -》 长度&&字符串
		// b.遍历allUser这个map，将map中的用户名发送给当前用户:key:=userid,将id和name拼接成一个字符串，返回给客户端
		userInput := string(buf[:n-1])
		if len(userInput) == 4 && userInput == "\\who" {
			// 遍历map
			fmt.Println("当前用户列表：")
			// 定义一个切片，包含所有用户信息
			var userInfos []string
			// 遍历map
			allusersMutex.RLock()
			for _, user := range allusers {
				userInfo := fmt.Sprintf("%s:%s\n", user.id, user.name)
				userInfos = append(userInfos, userInfo)
			}
			allusersMutex.RUnlock()
			//最终写到管道中，一定是个字符串
			r := strings.Join(userInfos, "\n")
			//将数据返回给查询的客户端
			newUser.msg <- r

		} else if len(userInput) > 9 && userInput[:7] == "\\rename" {
			// 2. rename
			// 规则:renane|Duke
			// 1. 获取用户输入,判断长度至少是7，判断字符是rename
			// 2. 获取用户输入,使用|进行切割，取第二个字段
			// 3. 更新用户名字newUser.name=Duke
			// 4. 告诉用户修改成功
			newName := strings.Split(userInput, "|")[1]
			newUser.name = newName

			allusersMutex.Lock()
			allusers[newUser.id] = newUser
			allusersMutex.Unlock()

			fmt.Println("修改用户名成功\n:", newUser.name)
			newUser.msg <- "修改用户名成功:" + newUser.name

		} else {
			if len(userInput) > 0 {
				// 将用户发送的消息广播给所有人
				msgInfo := fmt.Sprintf("[%s]: %s\n", newUser.name, userInput)
				message <- msgInfo
			}
		}
		// --------------------------------------------------------
	}
}

// 向所有的用户广播消息，启动一个全局的go程
func broadcast() {
	fmt.Println("广播启动")
	defer fmt.Println("广播结束,broadcast job结束")
	for {
		fmt.Println("广播中...")
		// 1.从message中读取数据
		info := <-message
		fmt.Println("广播的数据：", info)
		// 2.将数据写入到每一个msg管道中
		allusersMutex.RLock()
		for _, user := range allusers {
			user.msg <- info
		}
		allusersMutex.RUnlock()
	}
}

// 每个用户应该还有一个用来监听msg管道的go程，负责将数据返回客户端
func writeBackToClient(user *User, conn net.Conn) {
	defer func() {
		close(user.msg)
		fmt.Printf("用户 %s 的消息监听协程已退出\n", user.name)
	}()

	fmt.Printf("user: %s 的go程正在监听自己的msg管道:\n", user.name)
	// 1.从msg管道中读取数据
	for data := range user.msg {
		fmt.Printf("user:%s写会给客户端的数据为:%s\n", user.name, data)
		// 2.将数据返回给client
		_, err := conn.Write([]byte(data))
		if err != nil {
			fmt.Printf("向用户 %s 发送数据失败: %v\n", user.name, err)
			// 连接已关闭，退出循环
			// 通知handler协程退出
			select {
			case <-user.done:
				// 已经在退出过程中
			default:
				close(user.done)
			}
			break
		}
	}
}

// 启动一个go程，负责监听退出信号，触发后，进行清零工作：delete map，close conn都在这里
func watchExit(user *User, conn net.Conn, isQuit <-chan bool) {
	fmt.Println("watchExit...")
	defer fmt.Println("watchExit exit")
	for {
		select {
		case <-isQuit:
			logoutInfo := fmt.Sprintf("用户 %s 已退出聊天室", user.name)
			fmt.Println("删除当前用户", user.name)
			allusersMutex.Lock()
			delete(allusers, user.id)
			allusersMutex.Unlock()
			message <- logoutInfo
			// 通知handler协程退出
			close(user.done)
			fmt.Println("按任意键退出聊天框\n")
			return
		// 超时退出
		case <-time.After(time.Second * 40):
			logoutInfo := fmt.Sprintf("用户 %s 超时，已退出聊天室\n", user.name)
			fmt.Println("删除当前用户", user.name)
			allusersMutex.Lock()
			delete(allusers, user.id)
			allusersMutex.Unlock()
			message <- logoutInfo
			// 通知handler协程退出
			close(user.done)
			fmt.Println("按任意键退出聊天框\n")
			return
		}
	}
}
