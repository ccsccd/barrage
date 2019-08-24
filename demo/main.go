package main

import (
	"app/barrage"
	"github.com/gin-gonic/gin"
)

func main() {
	go func() {
		router := gin.Default()
		//主页
		router.StaticFile("/", "./static/screen.html")
		//弹幕系统
		hub := barrage.Hub{Connections: map[*barrage.WsConnection]bool{}}
		router.GET("/ws", hub.WsHandler)
		//随便抽一个幸运id，快，无参数
		router.GET("/lot", barrage.LotteryEnter)
		//随便抽一个说出口令的幸运id，慢，请求参数word="口令"
		router.GET("/red", barrage.RedPaccketEnter)
		//添加敏感词，请求参数word="敏感词"
		router.POST("/word", barrage.AddWord)
		_ = router.Run(":8080")
	}()
	barrage.Consumer()
}
