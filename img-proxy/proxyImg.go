package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	// 获取客户端的 IP 地址
	clientIP := r.RemoteAddr
	if strings.Contains(clientIP, ":") {
		// 如果 IP 地址是带端口的 (比如 [::1]:8080)，则去掉端口
		clientIP = strings.Split(clientIP, ":")[0]
	}

	// 获取请求 URL 参数
	path := r.URL.Path[len("/img-proxy/"):]

	// 判断是否有 URL 部分
	if path == "" {
		http.Error(w, "Missing URL to img-proxy", http.StatusBadRequest)
		return
	}

	// 修正 http:/ 或 https:/ 为 http:// 或 https://
	if strings.HasPrefix(path, "http:/") {
		path = "http:/" + path[5:]
	} else if strings.HasPrefix(path, "https:/") {
		path = "https:/" + path[6:]
	}

	// 解析目标 URL
	targetURL, err := url.Parse(path)
	if err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	// 打印访问日志：客户端 IP 和访问的 URL
	log.Printf("Client IP: %s requested: %s", clientIP, targetURL.String())

	// 发起请求到目标 URL
	resp, err := http.Get(targetURL.String())
	if err != nil {
		http.Error(w, "Failed to fetch the resource", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// 设置代理响应的头部和内容
	for k, v := range resp.Header {
		for _, vv := range v {
			w.Header().Add(k, vv)
		}
	}
	w.WriteHeader(resp.StatusCode)

	// 将远程资源的内容写入响应体
	io.Copy(w, resp.Body)
}

func main() {
	// 解析命令行参数
	port := flag.String("port", "8080", "Specify the port to listen on")
	flag.Parse()

	// 启动服务器
	http.HandleFunc("/proxy/", proxyHandler)
	fmt.Printf("Starting img-proxy server on :%s\n", *port)
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
