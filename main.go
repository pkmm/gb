package main

import (
	"github.com/pkmm/gb/baidu"
)

func main() {
	crawl, _ := baidu.NewBaiduTiebaCrawl("your bduss.")
	crawl.RunAtDaily()
	// 阻塞线程
	select {}
}
