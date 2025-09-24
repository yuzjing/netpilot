package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	// 一个简单的API处理器函数
	http.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
		// 设置响应头为JSON类型
		w.Header().Set("Content-Type", "application/json")
		// 允许前端跨域请求 (开发时非常重要!)
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// 写入一个简单的JSON响应
		fmt.Fprintln(w, `{"status": "ok", "message": "Welcome to NetPilot API!"}`)
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 允许前端跨域请求 (开发时非常重要!)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fmt.Fprintln(w, "pong")
	})

	fmt.Println("NetPilot API server is running on http://127.0.0.1:8080")
	// 启动HTTP服务器，监听8333端口
	log.Fatal(http.ListenAndServe("127.0.0.1:8080", nil))
}
