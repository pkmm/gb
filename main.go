package main

import (
	"fmt"
	"github.com/pkmm/gb/baidu"
)

func main() {
	//sample example
	worker := baidu.NewForumWorker("01T1AyRG1XY28ydTRNbVFTbmQzM1pwOXVMbnk2cWo3ODV1eFh6bW1DWXZvUjVhTVFBQUFBJCQAAAAAAAAAAAEAAABS5n44vrLLvNSwNAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAC8U91kvFPdZcW")
	data := baidu.ForumList{}
	data = append(data, baidu.Forum{"acm", worker.GetFid("acm")})
	ret := worker.SignAll(&data)
	fmt.Print(*ret)
}
