package main

import (
	"ad-data/router"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	//更新时间间隔,单位秒
	t time.Duration = 10
)

func main() {
	//定时监控数据
	//addata.TimingProject(t)//Web端已经有控制此函数功能

	gin.SetMode(gin.ReleaseMode)
	if err := router.NewRouter().Run(":8888"); err != nil {
		fmt.Println("Gin run failed:", err)
	}
}
