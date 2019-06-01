package main

import (
	"bufio"
	"fmt"
	"github.com/hpcloud/tail"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"
)


func main() {
	//文件监控通道，以行为单位写入到chan
	var ch = make(chan string,100)
	//下载结果chan，日志chan
	var result_log = make(chan string,100)
	//监控文件尾部获取url
	go watch_log(ch)
	//获取下载结果写入日志
	go log_write(result_log)
	for  {
		a := <- ch
		//fmt.Println(a)
		//从url的chan获取下载地址，下载结果写入日志chan
		go download(a,result_log)
	}
}

func log_write(ch chan string)  {
	f, err := os.OpenFile("result.log", os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println(err)
	}
	for  {
		a := <-ch
		//fmt.Println("write log")
		buf := []byte(a+"\n")
		f.Write(buf)
	}
}

func watch_log (ch chan string) {
	t, err := tail.TailFile("download.list", tail.Config{Follow: true})
	if err != nil {
		fmt.Println(err)
	}
	for line := range t.Lines {
		ch <- line.Text
	}
}

//func bytesToSize(length int) string {
//	var k = 1024 // or 1024
//	var sizes = []string{"Bytes", "KB", "MB", "GB", "TB"}
//	if length == 0 {
//		return "0 Bytes"
//	}
//	i := math.Floor(math.Log(float64(length)) / math.Log(float64(k)))
//	r := float64(length) / math.Pow(float64(k), i)
//	return strconv.FormatFloat(r, 'f', 3, 64) + " " + sizes[int(i)]
//}

func download(durl string,ch chan string)  {
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
	if resp.ContentLength <= 0 {
		log.Println("[*] Destination server does not support breakpoint download.")
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
	go func() {
		for {
			nr, er := reader.Read(buff)
			if nr > 0 {
				nw, ew := writer.Write(buff[0:nr])
				if nw > 0 {
					written += nw
				}
				if ew != nil {
					err = ew
					break
				}
				if nr != nw {
					err = io.ErrShortWrite
					break
				}
			}
			if er != nil {
				if er != io.EOF {
					err = er
				}
				break
			}
		}
		if err != nil {
			panic(err)
		}
	}()

	spaceTime := time.Second * 1
	ticker := time.NewTicker(spaceTime)
	lastWtn := 0
	stop := false
	fmt.Printf("%s is downloading\n",filename)

	for {
		select {
		case <-ticker.C:
			//speed := written - lastWtn
			//log.Printf("[*] Speed %s / %s \n", bytesToSize(speed), spaceTime.String())
			if written-lastWtn == 0 {
				ticker.Stop()
				stop = true
				break
			}
			lastWtn = written
		}
		if stop {
			ch <- fmt.Sprintf("%s is download done.",filename)
			fmt.Printf("%s is download finish.\n",filename)
			break
		}
	}
}
