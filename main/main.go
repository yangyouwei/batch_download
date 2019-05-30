package main

import (
	"fmt"
	"github.com/hpcloud/tail"
	"os"
	"path"
	"strings"
)

//url文件路径
var url_file string = "url.list"

//日志文件路径
var result_log string = "result.log"

//读取chanle
var chan_url = make(chan string, 100)

//日志写入chanle
var log_chan = make(chan string, 100)

func dowload(url string, logs chan string) {
	//文件下载
	u, n, p := cut_string(url)
	fmt.Println(u, n, p)
	logs <- u + n + p
}

func cut_string(str string) (url string, filenme string, list_path string) {
	//字符串分割  url 文件名 路径
	a := strings.Split(str, " ")
	filenme = path.Base(a[0])
	url = a[0]
	list_path = a[1]
	return
}

func watch(urls chan string) {
	t, _ := tail.TailFile(url_file, tail.Config{Follow: true})
	for line := range t.Lines {
		urls <- fmt.Sprint(line.Text)
	}
}

func log_write(result chan string) {
	a := <-result
	f, _ := os.OpenFile(result_log, os.O_WRONLY|os.O_APPEND, 0666)
	buf := []byte(a + "\n")
	f.Write(buf)
	f.Close()
}

func main() {
	go watch(chan_url)
	go log_write(log_chan)
	for {
		url := <-chan_url
		fmt.Println("for")
		go dowload(url, log_chan)
	}
}
