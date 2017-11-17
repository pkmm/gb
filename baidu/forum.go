package baidu

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
)

const (
	TBSURL   = "http://tieba.baidu.com/dc/common/tbs"
	FIDURL   = "http://tieba.baidu.com/mo/m"
	ILIKEURL = "http://tieba.baidu.com/mo/q-0--49BB6FC27D7013BD795602B74B2E83E2%3AFG%3D1--1-1-0----wapp_1462281637540_923/m?tn=bdFBW&tab=favorite"
	SIGNURL  = "http://c.tieba.baidu.com/c/c/forum/sign"
)

const (
	GET  = "GET"
	POST = "POST"
)

type ForumWorker struct {
	Cookie http.Cookie
	Client *http.Client
	Bduss  string
}

type Tbs struct {
	Tbs string
}

// 签到的forum的样子， 只需要2个标志
type Forum struct {
	Kw, Fid string
}

type ForumList []Forum

// 构造函数
func NewForumWorker(bduss string) *ForumWorker {
	return &ForumWorker{Cookie: http.Cookie{Name: "BDUSS", Value: bduss}, Client: &http.Client{}, Bduss: bduss}
}

// 配置请求的客户端
func (f ForumWorker) InitForumWorker() {

}

// 获取tbs
func (f ForumWorker) GetTbs() string {
	r, _ := http.NewRequest(GET, TBSURL, nil)
	r.AddCookie(&f.Cookie)
	resp, err := f.Client.Do(r)
	if err != nil {
		return ""
	}
	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	var tbs = Tbs{}
	json.Unmarshal(body, &tbs)
	return tbs.Tbs
}

// 获取fid
func (f ForumWorker) GetFid(kw string) string {
	args := url.Values{}
	args.Add("kw", kw)
	r, _ := http.NewRequest(GET, FIDURL+"?"+args.Encode(), nil)
	r.AddCookie(&f.Cookie)
	reply, err := f.Client.Do(r)
	if err != nil {
		return "-1"
	}
	defer reply.Body.Close()
	data, _ := ioutil.ReadAll(reply.Body)
	re, _ := regexp.Compile(`<input type="hidden" name="fid" value="(.*?)"/>`)
	match := re.FindSubmatch(data)
	if len(match) < 2 {
		return "-1"
	}
	return string(match[1])
}

// 获取关注的所有的贴儿吧
func (f ForumWorker) RetrieveForums() []string {
	r, _ := http.NewRequest(GET, ILIKEURL, nil)
	r.AddCookie(&f.Cookie)
	reply, err := f.Client.Do(r)
	if err != nil {
		return []string{}
	}
	defer reply.Body.Close()
	data, _ := ioutil.ReadAll(reply.Body)
	re, _ := regexp.Compile(`<a href=".*?kw=.*?">(.*?)</a>`)
	machs := re.FindAllSubmatch(data, -1)
	var ret []string
	for _, name := range machs {
		if len(name) < 2 {
			continue
		}
		ret = append(ret, string(name[1]))
	}
	return ret
}

// 签到一个贴吧
func (f ForumWorker) signOne(kw, fid string, ch chan string) {
	formData := url.Values{
		"BDUSS":           {f.Bduss},
		"_client_id":      {"03-00-DA-59-05-00-72-96-06-00-01-00-04-00-4C-43-01-00-34-F4-02-00-BC-25-09-00-4E-36"},
		"_client_type":    {"4"},
		"_client_version": {"1.2.1.17"},
		"_phone_imei":     {"540b43b59d21b7a4824e1fd31b08e9a6"},
		"fid":             {fid},
		"kw":              {kw},
		"net_type":        {"3"},
		"tbs":             {f.GetTbs()},
	}
	formData = f.encrypt(formData)
	r, _ := http.NewRequest("POST", SIGNURL, strings.NewReader(formData.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("User-Agent", "Pkmm iPhone/1.0 BadApple/99.1")
	r.AddCookie(&f.Cookie)
	resp, _ := f.Client.Do(r)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	ch <- kw
	ch <- string(body)
}

// 规格化上传的数据
func (f ForumWorker) encrypt(formData url.Values) url.Values {
	var keys = make([]string, len(formData))
	var i = 0
	for k := range formData {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	var encrypts bytes.Buffer
	for _, k := range keys {
		encrypts.WriteString(k + "=" + formData.Get(k))
	}
	encrypts.WriteString("tiebaclient!!!")
	md5value := md5.New()
	md5value.Write(encrypts.Bytes())
	sign := hex.EncodeToString(md5value.Sum(nil))
	sign = strings.ToUpper(sign)
	formData.Set("sign", sign)
	return formData
}

// 签到所有贴吧，使用goroutines
func (f ForumWorker) SignAll(needSignForums *ForumList) *map[string]string {
	size := len(*needSignForums)
	localChannel := make(chan string, size<<1)
	for _, forum := range *needSignForums {
		go f.signOne(forum.Kw, forum.Fid, localChannel)
	}
	result := make(map[string]string, size)
	for i := 0; i < size; i++ {
		kw := <-localChannel
		signResult := <-localChannel
		result[kw] = signResult
		fmt.Println(signResult)
	}
	return &result
}

// many days ago

//var BDUSS = ""
//var client = &http.Client{}
//var cookie = http.Cookie{Name: "BDUSS", Value: BDUSS}
//func GetTbs() string {
//	req, _ := http.NewRequest("GET", TBS_URL, nil)
//	req.AddCookie(&cookie)
//	rep, err := client.Do(req)
//	data, _ := ioutil.ReadAll(rep.Body)
//	defer rep.Body.Close()
//	type TBS struct {
//		Tbs string
//	}
//	var tbs TBS
//	if err == nil {
//		json.Unmarshal(data, &tbs)
//		return tbs.Tbs
//	}
//	return ""
//}
//func GetFid(kw string) string {
//	args := url.Values{}
//	args.Add("kw", kw)
//	request, _ := http.NewRequest("GET", FID_URL+"?"+args.Encode(), nil)
//	request.AddCookie(&cookie)
//	rep, err := client.Do(request)
//	if err != nil {
//		return "-1"
//	}
//	defer rep.Body.Close()
//	data, _ := ioutil.ReadAll(rep.Body)
//	re, _ := regexp.Compile(`<input type="hidden" name="fid" value="(.*?)"/>`)
//	submatch := re.FindSubmatch(data)
//	if len(submatch) < 2 {
//		return "-1"
//	}
//	return string(submatch[1])
//}
//func GetAllstar() []string {
//	req, _ := http.NewRequest("GET", I_LIKE_URL, nil)
//	req.AddCookie(&cookie)
//	rep, err := client.Do(req)
//	if err != nil {
//	}
//	defer rep.Body.Close()
//	data, _ := ioutil.ReadAll(rep.Body)
//	re, _ := regexp.Compile(`<a href=".*?kw=.*?">(.*?)</a>`)
//	submatch := re.FindAllSubmatch(data, -1)
//	ret := []string{}
//	for _, name := range submatch {
//		if len(name) < 2 {
//			continue
//		}
//		ret = append(ret, string(name[1]))
//	}
//	return ret
//}
//func Sign(kw, fid string, ch chan string) {
//	data := url.Values{
//		"BDUSS":           {BDUSS},
//		"_client_id":      {"03-00-DA-59-05-00-72-96-06-00-01-00-04-00-4C-43-01-00-34-F4-02-00-BC-25-09-00-4E-36"},
//		"_client_type":    {"4"},
//		"_client_version": {"1.2.1.17"},
//		"_phone_imei":     {"540b43b59d21b7a4824e1fd31b08e9a6"},
//		"fid":             {fid},
//		"kw":              {kw},
//		"net_type":        {"3"},
//		"tbs":             {GetTbs()},
//	}
//	var keys = make([]string, len(data))
//	var i = 0
//	for k, _ := range data {
//		keys[i] = k
//		i++
//	}
//	sort.Strings(keys)
//	var enerypt bytes.Buffer
//	for _, k := range keys {
//		enerypt.WriteString(k)
//		enerypt.WriteString("=")
//		enerypt.WriteString(data.Get(k))
//	}
//	enerypt.WriteString("tiebaclient!!!")
//	tt := md5.New()
//	tt.Write(enerypt.Bytes())
//	sign := hex.EncodeToString(tt.Sum(nil))
//	sign = strings.ToUpper(sign)
//	data.Set("sign", sign)
//	req, _ := http.NewRequest("POST", SIGN_URL, strings.NewReader(data.Encode()))
//	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
//	req.Header.Add("User-Agent", "Pkmm iPhone/1.0 BadApple/99.1")
//	req.AddCookie(&cookie)
//	rep, _ := client.Do(req)
//	defer rep.Body.Close()
//	html, _ := ioutil.ReadAll(rep.Body)
//	fmt.Println(string(html))
//	type ResponseJson struct {
//		Error_code string
//		Error_msg  string
//		Info       []string
//		Error      map[string]string
//	}
//	ret := ResponseJson{}
//	json.Unmarshal(html, &ret)
//	ch <- string(html)
//	ch <- kw
//}
