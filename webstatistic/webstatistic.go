package webstatistic

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	// http升级websocket协议的配置
	wsUpgrader = websocket.Upgrader{
		// 允许所有CORS跨域请求
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	//rwWDM 用于存储各个网站的在线用户量
	rwWDM = struct {
		sync.RWMutex
		data map[string]int
	}{
		data: make(map[string]int),
	}
)

func main() {
	go ClearMap()
	gin.SetMode(gin.ReleaseMode)
	Router := gin.Default()
	Router.GET("/ws/onlineServer", func(c *gin.Context) {
		if c.Query("weburl") == "" {
			c.JSON(200, gin.H{
				"msg": "滚！！！",
			})
			return
		}
		WebSocketHandler(c.Query("weburl"), c.Writer, c.Request)
	})
	if err := Router.Run(":8899"); err != nil {
		fmt.Println("Gin run failed:", err)
	}
}

//WebSocketHandler /ws/onlineServer接口函数
func WebSocketHandler(clientURL string, resp http.ResponseWriter, req *http.Request) {
	// 应答客户端告知升级连接为websocket
	wsSocket, err := wsUpgrader.Upgrade(resp, req, nil)
	if err != nil {
		return
	}
	rwWDM.Lock()
	rwWDM.data[clientURL]++
	rwWDM.Unlock()
	//结束后关闭websocket连接
	defer func() {
		_ = wsSocket.Close()
	}()
	//发送在线人数给客户端，相当于1秒检测一遍用户是否还在
	for {
		if err = wsSocket.WriteMessage(websocket.TextMessage, getNewestNum(clientURL)); err != nil {
			rwWDM.Lock()
			rwWDM.data[clientURL]--
			rwWDM.Unlock()
			return
		}
		time.Sleep(time.Second * 1)
	}
}

//ClearMap 协程，用于重置map
func ClearMap() {
	for {
		rwWDM.Lock()
		rwWDM.data = make(map[string]int)
		rwWDM.Unlock()
		time.Sleep(time.Minute * 60)
	}
}

//getNewestNum 获取最新在线人数
func getNewestNum(k string) []byte {
	rwWDM.RLock()
	n := rwWDM.data[k]
	rwWDM.RUnlock()
	return []byte(strconv.Itoa(n))
}
