package main

import (
	B "Baidu/baidu"
	mysqlHelper "Baidu/storage/mysql"
	"fmt"
	"pkmm/utils/baidu"
)

func getFid(kw, bduss string, ch chan string) {
	fid := B.GetFid(kw, bduss)
	ch <- kw
	ch <- fid
}

func SyncByUserId(userId, bduss string) { // 同步指定用户的贴吧
	kws := B.GetAllstar(bduss)
	var N = len(kws)
	var ch = make(chan string, N*2)
	for _, kw := range kws {
		go getFid(kw, bduss, ch)
	}
	var data = make(map[string]string)
	var i = 0
	for i < N {
		i++
		kw := <-ch
		fid := <-ch
		data[kw] = fid
	}
	mysqlHelper.SyncFromOffical(&data, userId)
}

func SignByUserId(userId, bduss string) {
	data := mysqlHelper.Get(userId)
	N, i := len(*data), 0
	ch := make(chan string, N*2)
	for kw, fid := range *data {
		go B.Sign(kw, fid, bduss, ch)
	}
	ret := make(map[string]mysqlHelper.Info)
	for i < N {
		i++
		var kw, infos string
		infos = <-ch
		kw = <-ch
		fmt.Println(kw)
		fmt.Println(infos)
		ret[kw] = mysqlHelper.Info{(*data)[kw], infos}
	}
	mysqlHelper.Update(&ret, "1")
}

func main() {
<<<<<<< HEAD
	t := baidu.NewForunWorker("01T1AyRG1XY28ydTRNbVFTbmQzM1pwOXVMbnk2cWo3ODV1eFh6bW1DWXZvUjVhTVFBQUFBJCQAAAAAAAAAAAEAAABS5n44vrLLvNSwNAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAC8U91kvFPdZcW")
	//fmt.Println(t.GetTbs())
	//fmt.Println(t.GetFid("浙江中医药大学"))
	//fmt.Println(t.GetAllForums())
	data := baidu.ForumList{}
	data = append(data, baidu.Forum{"acm", t.GetFid("acm")})

	fmt.Println(*t.SignAll(&data))
=======
	var bduss = "" // bduss
	SyncByUserId("1", bduss)
	SignByUserId("1", bduss)
>>>>>>> 0a0521446b31b34457598c8c4f05123d4e303ca4
}
