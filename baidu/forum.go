package baidu

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	tbsUrl   = "http://tieba.baidu.com/dc/common/tbs"
	fidUrl   = "http://tieba.baidu.com/mo/m"
	iLikeUrl = "http://tieba.baidu.com/mo/q-0--49BB6FC27D7013BD795602B74B2E83E2%3AFG%3D1--1-1-0----wapp_1462281637540_923/m?tn=bdFBW&tab=favorite"
	signUrl  = "http://c.tieba.baidu.com/c/c/forum/sign"
)

const (
	GET  = "GET"
	POST = "POST"
)

type Tbs struct {
	Tbs string
}

// 签到某一个贴吧需要的两个参数
type Forum struct {
	Kw, Fid string
}

type ForumList []Forum

type Crawl struct {
	client *http.Client
	cookie *http.Cookie
	bduss  string
}

func NewBaiduTiebaCrawl(bduss string) (*Crawl, error) {
	timeout := time.Duration(30 * time.Second)
	client := &Crawl{bduss: bduss}
	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		return nil, errors.New("init cookieJar failed => " + err.Error())
	}
	client.client = &http.Client{
		Timeout: timeout,
		Jar:     cookieJar,
	}
	client.cookie = &http.Cookie{Name:"BDUSS", Value:bduss}
	return client, nil
}

func (c *Crawl) SetBduss(bduss string) *Crawl {
	c.bduss = bduss
	c.cookie = &http.Cookie{Name:"BDUSS", Value:bduss}
	return c
}

func (c *Crawl) getTbs() string {
	r, err := http.NewRequest(GET, tbsUrl, nil)
	if err != nil {
		return ""
	}
	r.AddCookie(c.cookie)
	resp, err := c.client.Do(r)
	if err != nil {
		return ""
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	var tbs = Tbs{}
	json.Unmarshal(body, &tbs)
	return tbs.Tbs
}

func (c *Crawl) getFid(kw string) string {
	args := url.Values{}
	args.Add("kw", kw)
	r, _ := http.NewRequest(GET, fidUrl+"?"+args.Encode(), nil)
	r.AddCookie(c.cookie)
	reply, err := c.client.Do(r)
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
func (c *Crawl) RetrieveForums() []string {
	r, _ := http.NewRequest(GET, iLikeUrl, nil)
	r.AddCookie(c.cookie)
	reply, err := c.client.Do(r)
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
func (c *Crawl) signOne(kw, fid string, ch chan string) {
	formData := url.Values{
		"BDUSS":           {c.bduss},
		"_client_id":      {"03-00-DA-59-05-00-72-96-06-00-01-00-04-00-4C-43-01-00-34-F4-02-00-BC-25-09-00-4E-36"},
		"_client_type":    {"4"},
		"_client_version": {"1.2.1.17"},
		"_phone_imei":     {"540b43b59d21b7a4824e1fd31b08e9a6"},
		"fid":             {fid},
		"kw":              {kw},
		"net_type":        {"3"},
		"tbs":             {c.getTbs()},
	}
	formData = c.encrypt(formData)
	r, _ := http.NewRequest("POST", signUrl, strings.NewReader(formData.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("User-Agent", "Pkmm iPhone/1.0 BadApple/99.1")
	r.AddCookie(c.cookie)
	resp, err := c.client.Do(r)
	if err != nil {
		ch <- ""
		ch <- ""
		return
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	ch <- kw
	ch <- string(body)
}

// 规格化上传的数据
func (c *Crawl) encrypt(formData url.Values) url.Values {
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
func (c *Crawl) SignAll(needSignForums *ForumList) *map[string]string {
	size := len(*needSignForums)
	localChannel := make(chan string, size<<1)
	for _, forum := range *needSignForums {
		go c.signOne(forum.Kw, forum.Fid, localChannel)
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
