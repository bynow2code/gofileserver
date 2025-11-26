package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	dir := flag.String("dir", ".", "要共享的目录（默认当前目录）")
	port := flag.String("port", "8080", "监听端口（默认8080）")
	flag.Parse()

	absDir, err := filepath.Abs(*dir)
	if err != nil {
		fmt.Printf("路径解析错误：[%s]\n", err)
		os.Exit(1)
	}
	_, err = os.Stat(absDir)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("指定的目录不存在：[%s]\n", err)
			os.Exit(1)
		} else {
			fmt.Printf("访问目录失败：[%s]\n", err)
			os.Exit(1)
		}
	}

	fileServer := http.FileServer(http.Dir(absDir))
	http.Handle("/", fileServer)

	fmt.Printf("✅ 文件服务器已启动！\n")
	fmt.Printf("📂 共享目录：%s\n", absDir)
	fmt.Printf("🌐 访问地址：http://localhost:%s\n", *port)

	err = http.ListenAndServe(":"+(*port), nil)
	if err != nil {
		fmt.Printf("文件服务启动失败：[%s]\n", err)
		os.Exit(1)
	}

}
