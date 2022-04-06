package main

import (
	"bytes"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type ChContent struct {
	Content []byte
	File    []byte
	Status  int //1=Content 2=File
	Pid     int
}

var tmp ChContent
var LsbRelease string

func Xclip(ch chan ChContent) {
	x3 := exec.Command("lsb_release", "-is")
	lsb, _ := x3.Output()
	LsbRelease = string(lsb)
	//log.Println(LsbRelease)
	for {
		tmp.Status = 0
		primary := exec.Command("xclip", "-o", "-selection", "primary")
		primaryByte, _ := primary.Output()
		clipboard := exec.Command("xclip", "-o", "-selection", "clipboard")
		clipboardByte, _ := clipboard.Output()

		if len(primaryByte) > 0 || len(clipboardByte) > 0 {
			cmd := exec.Command("sh", "-c", "xprop -id $(xprop -root | awk '/_NET_ACTIVE_WINDOW\\(WINDOW\\)/{print $NF}') | awk '/_NET_WM_PID\\(CARDINAL\\)/{print $NF}'")
			bytes, _ := cmd.Output()
			pid, _ := strconv.Atoi(strings.Replace(string(bytes), "\n", "", -1)) //strings.Replace(string(bytes), "\n", " ", -1)
			tmp.Pid = pid
			if string(primaryByte) == string(clipboardByte) {
				if string(tmp.Content) != string(primaryByte) {
					tmp.Content = primaryByte
					tmp.Status = 1
				}

			} else {
				if string(tmp.Content) != string(clipboardByte) && string(tmp.File) != string(clipboardByte) {
					tmp.File = clipboardByte
					tmp.Status = 2
				}

			}
			if tmp.Status > 0 {
				ch <- tmp
			}

		}

		if strings.Replace(LsbRelease, "\n", "", -1) == "NFS Desktop" {
			//log.Println("lsb_release -is:", LsbRelease)
			time.Sleep(time.Millisecond * 1000)
		} else {
			time.Sleep(time.Millisecond * 50)
		}

	}

}

func main() {

	ch := make(chan ChContent, 1)
	go Xclip(ch)

	for {
		//<-ch
		str := <-ch

		logStr := strconv.Itoa(str.Status) + "|" + strconv.Itoa(str.Pid)
		if str.Status == 1 { //复制文本
			logStr = logStr + "|" + string(str.Content) //日志

			//log.Println("copy content :(", (str.Pid), ") | ", string(str.Content), "\n")
			log.Println("|copy content|", logStr)
		} else if str.Status == 2 { //复制内容
			logStr = logStr + "|" + string(bytes.ReplaceAll(str.File, []byte("\n"), []byte("|"))) //日志
			var byteFile [][]byte
			if strings.Replace(LsbRelease, "\n", "", -1) == "Ubuntu" {
				ubuntuFile := bytes.Split(str.File, []byte("\n"))
				i := 0
				//log.Println("len ubuntuFile", len(ubuntuFile)-1)
				byteFile = make([][]byte, len(ubuntuFile)-2)
				for _, val := range ubuntuFile {
					//log.Println(string(val[:7]), len(val))
					if len(val) > 7 && string(val[:7]) == "file://" {
						byteFile[i] = val
						i++
					}

				}
				//
			} else {
				byteFile = bytes.Split(str.File, []byte("\n"))
			}

			//对 隐藏文件过滤
			var list []string //make([]string, 0)
			for _, file := range byteFile {
				split := bytes.Split(file, []byte("/"))

				if len(file) > 0 && string(split[len(split)-1][0:1]) != "." {
					list = append(list, string(file[:]))
					//log.Println("end", string(file), string(split[len(split)-1]))
				}

			}
			//log.Println("------", list)
			log.Println("|copy file|", logStr) // strings.Replace(str.Content, "\n", " ", -1)
		}

	}
}
