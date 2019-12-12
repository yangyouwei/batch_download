package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"sync"
	"time"
)


func main() {
	durl := "http://mirrors.cn99.com/centos/8.0.1905/isos/x86_64/CentOS-8-x86_64-1905-dvd1.iso"
	download(durl)
}

func download(durl string)  {
	wg := sync.WaitGroup{}
	uri, err := url.ParseRequestURI(durl)
	if err != nil {
		panic("网址错误")
	}

	filename := path.Base(uri.Path)
	log.Println("[*] Filename " + filename)

	client := http.DefaultClient;
	client.Timeout = time.Second * 3600 //设置超时时间
	resp, err := client.Get(durl)
	if err != nil {
		panic(err)
	}

	raw := resp.Body
	defer raw.Close()
	reader := bufio.NewReaderSize(raw, 1024*32);

	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	writer := bufio.NewWriter(file)

	buff := make([]byte, 32*1024)
	written := 0
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		for {
			nr, er := reader.Read(buff)
			if nr > 0 {
				nw, ew := writer.Write(buff[0:nr])
				if nw > 0 {
					written += nw //累加下载数据值
				}
				if ew != nil { // 读取缓存，写入文件失败，可能是写完了。
					fmt.Println(filename," ",ew)
					break
				}
				if nr != nw {	//读取和写入的数值不相等，写入不完整，下载失败
					err = io.ErrShortWrite
					fmt.Println(filename," ",err)
					break
				}
			}
			if er != nil {	   //错误或者完毕都要跳出循环
				if er != io.EOF {  //是错误，不是完毕，打印错误信息
					err = er
					fmt.Println(filename," aaa",err)  //打印错误
				}
				fmt.Sprintf("%s is download finish.",filename)
				break
			}
		}
		wg.Done()
	}(&wg)
	wg.Wait()
}
