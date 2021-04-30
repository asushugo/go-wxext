package main

import (
	"fmt"
	"github.com/asushugo/go-wxext/wxext"
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
				data := m["data"].(map[string]interface{})
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

	wx := wxext.NewWxext(
		"go_wxext",
		"ABDCE3C5429E3045076F1CE232C40B21",
		// 如果你反向代理Websocket，可能就需要设置以下选项，否则默认注释即可
		//wxext.SetAddr("192.168.2.212"),
		//wxext.SetPort(82),
		//wxext.SetWebsocketPort(81),
	)

	err := wx.Conn()
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
	wx.List()

	// 监听退出信号
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("get signal: %s", <-c)
	}()
	log.Println(<-errChan)
}
