package main

import (
	"fmt"
	"github.com/pkmm/gb/baidu"
)

func main() {
	//sample example
	worker := baidu.NewForumWorker("your baidu bduss")
	data := baidu.ForumList{}
	data = append(data, baidu.Forum{"贴吧名称", "贴吧fid"})
	ret := worker.SignAll(&data)
	fmt.Print(*ret)
}
