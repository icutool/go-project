package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
)

func getClientIP(r *http.Request) string {
	// 尝试从 X-Real-IP 和 X-Forwarded-For 头部获取真实 IP
	clientIP := r.Header.Get("X-Real-IP")
	if clientIP == "" {
		clientIP = r.Header.Get("X-Forwarded-For")
	}
	if clientIP == "" {
		// 如果没有找到头部，就退回到 RemoteAddr
		clientIP, _, _ = net.SplitHostPort(r.RemoteAddr)
	}
	return clientIP
}

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	// 获取客户端的 IP 地址
	clientIP := getClientIP(r)

	// 获取请求 URL 参数
	path := r.URL.Path[len("/proxy/"):]
	if path == "" {
		http.Error(w, "Missing URL to proxy", http.StatusBadRequest)
		return
	}

	// 修正 http:/ 或 https:/ 为 http:// 或 https://
	if strings.HasPrefix(path, "http:/") {
		path = "http:/" + path[5:]
	} else if strings.HasPrefix(path, "https:/") {
		path = "https:/" + path[6:]
	}

	// 获取查询字符串部分
	queryString := r.URL.RawQuery
	if queryString != "" {
		path = path + "?" + queryString
	}

	// 解析目标 URL
	targetURL, err := url.Parse(path)
	if err != nil {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	// 解码 URL
	decodedPath, err := url.QueryUnescape(targetURL.String())
	if err != nil {
		http.Error(w, "Failed to decode the URL", http.StatusBadRequest)
		return
	}
	targetURL, err = url.Parse(decodedPath)
	if err != nil {
		http.Error(w, "Invalid decoded URL", http.StatusBadRequest)
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
	fmt.Printf("Starting proxy server on :%s\n", *port)
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
