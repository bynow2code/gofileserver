package main

import (
	"embed"
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
)

var version = "0.0.0-dev"

//go:embed templates/*
var tplFS embed.FS

// main 函数是程序的入口点，用于启动一个简单的HTTP文件服务器。
// 它通过命令行参数接收要共享的目录和监听端口，并提供该目录下文件的浏览和下载功能。
//
// 参数说明：
//
//	-dir string
//	   要共享的目录路径，默认为当前目录(".")
//	-port string
//	   HTTP服务器监听的端口号，默认为"8080"
//
// 返回值：
//
//	无返回值，但会根据运行情况调用os.Exit退出程序
func main() {
	dir := flag.String("dir", ".", "要共享的目录（默认当前目录）")
	port := flag.String("port", "8080", "监听端口（默认8080）")
	flag.Parse()

	// 解析并验证目录路径的有效性
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

	// 注册根路径处理函数，用于处理所有HTTP请求
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 根据请求路径构建实际文件系统路径
		realPath := filepath.Join(absDir, r.URL.Path)
		file, err := os.Stat(realPath)
		if err != nil {
			http.Error(w, fmt.Sprintf("路径不存在：[%s]", err), http.StatusNotFound)
			return
		}

		if !file.IsDir() {
			// 处理文件请求，直接提供文件内容
			http.ServeFile(w, r, realPath)
			return
		} else {
			// 处理目录请求，渲染目录列表页面
			renderDirList(w, r, realPath)
		}
	})

	// 输出服务器启动信息
	fmt.Printf("✅ 文件服务器已启动！\n")
	fmt.Printf("📂 共享目录：%s\n", absDir)
	fmt.Printf("🌐 访问地址：http://localhost:%s\n", *port)

	// 启动HTTP服务器开始监听请求
	err = http.ListenAndServe(":"+(*port), nil)
	if err != nil {
		fmt.Printf("文件服务启动失败：[%s]\n", err)
		os.Exit(1)
	}
}

// renderDirList 渲染目录列表页面
// 参数:
//
//	w: HTTP响应写入器，用于向客户端发送响应
//	r: HTTP请求对象，包含客户端请求信息
//	realPath: 实际文件系统路径，表示要列出的目录位置
func renderDirList(w http.ResponseWriter, r *http.Request, realPath string) {
	// 读取指定路径下的目录内容
	fi, err := os.ReadDir(realPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("读取目录错误：[%s]", err), http.StatusInternalServerError)
		return
	}

	// 构造文件信息列表，包含文件名、是否为目录、访问URL等信息
	var files []map[string]any
	for _, file := range fi {
		files = append(files, map[string]any{
			"name":  file.Name(),
			"isDir": file.IsDir(),
			"url":   filepath.Join(r.URL.Path, file.Name()),
		})
	}

	// 获取上级目录URL路径
	parentURL := filepath.Dir(r.URL.Path)

	// 解析HTML模板文件
	tmpl, err := template.ParseFS(tplFS, "templates/template.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("解析模板错误：[%s]", err), http.StatusInternalServerError)
		return
	}

	// 执行模板渲染，将数据填充到模板中并输出到HTTP响应
	err = tmpl.Execute(w, map[string]any{
		"currentPath": r.URL.Path,
		"files":       files,
		"parentURL":   parentURL,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("执行模版错误：[%s]", err), http.StatusInternalServerError)
		return
	}
}
