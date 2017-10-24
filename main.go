package main

import (
	B "Baidu/baidu"
	mysqlHelper "Baidu/storage/mysql"
	"fmt"
)

func getFid(kw string, ch chan string) {
	fid := B.GetFid(kw)
	ch <- kw
	ch <- fid
}

func SyncByUserId(userId string) { // 同步指定用户的贴吧
	kws := B.GetAllstar()
	var N = len(kws)
	var ch = make(chan string, N*2)
	for _, kw := range kws {
		go getFid(kw, ch)
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

func SignByUserId(userId string) {
	data := mysqlHelper.Get(userId)
	N, i := len(*data), 0
	ch := make(chan string, N*2)
	for kw, fid := range *data {
		go B.Sign(kw, fid, ch)
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
	SyncByUserId("1")
	SignByUserId("1")
}
