package main

import (
	"flag"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
)

var version = "0.0.0-dev"

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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		realPath := filepath.Join(absDir, r.URL.Path)
		file, err := os.Stat(realPath)
		if err != nil {
			http.Error(w, fmt.Sprintf("路径不存在：[%s]", err), http.StatusNotFound)
			return
		}

		if !file.IsDir() {
			// 文件
			http.ServeFile(w, r, realPath)
			return
		} else {
			// 目录
			renderDirList(w, r, realPath)
		}
	})

	fmt.Printf("✅ 文件服务器已启动！\n")
	fmt.Printf("📂 共享目录：%s\n", absDir)
	fmt.Printf("🌐 访问地址：http://localhost:%s\n", *port)

	err = http.ListenAndServe(":"+(*port), nil)
	if err != nil {
		fmt.Printf("文件服务启动失败：[%s]\n", err)
		os.Exit(1)
	}
}

func renderDirList(w http.ResponseWriter, r *http.Request, realPath string) {
	fi, err := os.ReadDir(realPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("读取目录错误：[%s]", err), http.StatusInternalServerError)
		return
	}

	var files []map[string]any
	for _, file := range fi {
		files = append(files, map[string]any{
			"name":  file.Name(),
			"isDir": file.IsDir(),
			"url":   filepath.Join(r.URL.Path, file.Name()),
		})
	}

	parentURL := filepath.Dir(r.URL.Path)

	tmpl, err := template.ParseFiles("template.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("解析模板错误：[%s]", err), http.StatusInternalServerError)
		return
	}

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
