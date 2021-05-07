package main

import (
	"fmt"
	"github.com/asushugo/go-wxext/wxext"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

func MsgHandler(wx *wxext.Wxext) {
	for recv := range wx.Recv {

		// 开协程处理信息
		go func(m map[string]interface{}) {
			method := m["method"].(string)
			// 更多事件请参考
			// https://www.wxext.cn/home/developer.html#/?id=%e4%ba%8b%e4%bb%b6%e9%80%9a%e7%9f%a5
			if strings.Index(method, "Recv") == -1 { // 消息事件
				data, ok := m["data"].(map[string]interface{})
				if ok != true {
					log.Println(m)
					return
				}
				switch method {
				case "newmsg": // 微信消息事件
					typeId := int(m["type"].(float64))
					switch typeId {
					case wxext.TextMessage:
						msg := data["msg"].(string)
						fromid := data["fromid"].(string)
						log.Println(fmt.Sprintf("群组：%s\t成员：%s\t消息：%s", fromid, data["memid"].(string), msg))
						// 群复读机
						if fromid == "24614778494@chatroom" {
							wx.SendText(int32(m["pid"].(float64)), "24614778494@chatroom", "", msg)
						}
					case wxext.ImgMessage: // 照片事件
						log.Println(data)
					}
				case "xmlinfo": // xml事件
					typeId := int(m["type"].(float64))
					switch typeId {
					case wxext.XMLImgPathMessage: // 照片路径事件
						log.Println(data["path"])
					}
				default: // 其他事件
					//log.Println(m)
				}
			} else { // 函数事件
				switch method {
				case "getInfo_Recv": // 执行getInfo的事件
					log.Println(m["data"].(map[string]interface{}))
				case "list_Recv": // method: list
					log.Println(m)
				default:
				}
			}
		}(recv)
	}
}

func main() {
	errChan := make(chan error)

	// 输出日志
	logFile, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer logFile.Close()
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)

	// 从环境变量获取Name和Key
	env := os.Getenv("cn.wxext.app")
	if env == "" {
		log.Fatal("环境变量为空！")
	}
	envList := strings.Split(env, ";")

	log.Println(fmt.Sprintf("插件名称：%s", envList[0]))
	log.Println(fmt.Sprintf("Key：%s", envList[1]))

	//wx := wxext.NewWxext(
	//	"go_wxext",
	//	"ABDCE3C5429E3045076F1CE232C40B21",
	//	//如果你反向代理Websocket，可能就需要设置以下选项，否则默认注释即可
	//	//wxext.SetAddr("192.168.2.212"),
	//	//wxext.SetPort(82),
	//	//wxext.SetWebsocketPort(81),
	//)

	wx := wxext.NewWxext(
		envList[0], // Name
		envList[1], // Key
	)

	err = wx.Conn()
	if err != nil {
		log.Fatal(err)
	}

	// 如果KEY不正确或者Wxext关闭都会导致连接中断
	go func() {
		for range wx.ErrChan {
			log.Fatal("连接中断！")
		}
	}()

	go MsgHandler(wx)
	time.Sleep(1 * time.Second)

	log.Println("running......")
	wx.List() // 获取连接列表

	// 监听退出信号
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("get signal: %s", <-c)
	}()
	log.Println(<-errChan)
}
