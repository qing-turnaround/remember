package main

import (
	"fmt"
	"remember/markdown"
	"remember/mysql"
	"remember/settings"
	"time"
)

var (
	local = time.Now()
)

func main() {
	// 初始化配置
	if err := settings.Init("config.yaml"); err != nil {
		fmt.Println("初始配置错误", err.Error())
	}

	if err := mysql.Init(); err != nil {
		fmt.Println("mysql 初始化失败", err.Error())
	}
	defer mysql.Close()

	markdown.Logic(local)
}

