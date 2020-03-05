package baidu

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
)

// 贴吧签到的结果
// 成功的信息 失败的信息
type tiebaSignInfo struct {
	ErrorCode  string `json:"error_code"`
	ErrorMsg   string `json:"error_msg"`
	Time       int    `json:"time,omitempty"`
	Ctime      int    `json:"ctime,omitempty"`
	Logid      int    `json:"logid,omitempty"`
	ServerTime string `json:"server_time,omitempty"`
}

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

type Tieba struct {
	Kw, Fid string
}

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
	client.cookie = &http.Cookie{Name: "BDUSS", Value: bduss}
	return client, nil
}

func (c *Crawl) SetBduss(bduss string) *Crawl {
	c.bduss = bduss
	c.cookie = &http.Cookie{Name: "BDUSS", Value: bduss}
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
	r, err := http.NewRequest(GET, fidUrl+"?"+args.Encode(), nil)
	if err != nil {
		return "-1"
	}
	r.AddCookie(c.cookie)
	reply, err := c.client.Do(r)
	if err != nil {
		return "-1"
	}
	defer reply.Body.Close()
	data, err := ioutil.ReadAll(reply.Body)
	if err != nil {
		return "-1"
	}
	re, err := regexp.Compile(`<input type="hidden" name="fid" value="(.*?)"/>`)
	if err != nil {
		return "-1"
	}
	match := re.FindSubmatch(data)
	if len(match) < 2 {
		return "-1"
	}
	return string(match[1])
}

func (c *Crawl) RetrieveTiebas() ([]string, error) {
	r, err := http.NewRequest(GET, iLikeUrl, nil)
	if err != nil {
		return nil, err
	}
	r.AddCookie(c.cookie)
	reply, err := c.client.Do(r)
	if err != nil {
		return nil, err
	}
	defer reply.Body.Close()
	data, err := ioutil.ReadAll(reply.Body)
	if err != nil {
		return nil, err
	}
	re, err := regexp.Compile(`<a href=".*?kw=.*?">(.*?)</a>`)
	if err != nil {
		return nil, err
	}
	machs := re.FindAllSubmatch(data, -1)
	var ret []string
	for _, name := range machs {
		if len(name) < 2 {
			continue
		}
		ret = append(ret, string(name[1]))
	}
	return ret, nil
}

type signResult struct {
	kw   string // 签到的贴吧名称
	resp string // 签到的响应结果
}

// 签到一个贴吧
// kw: 贴吧名称
// 返回签到结果的json string | ""
func (c *Crawl) SignOne(kw string) string {
	fid := c.getFid(kw)
	if fid == "-1" {
		return ""
	}
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
	r, err := http.NewRequest("POST", signUrl, strings.NewReader(formData.Encode()))
	if err != nil {
		return ""
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("User-Agent", "Pkmm iPhone/1.0 BadApple/99.1")
	r.AddCookie(c.cookie)
	resp, err := c.client.Do(r)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	return string(body)
}

// 签到一个贴吧
// kw: 贴吧名称
func (c *Crawl) signOne(kw string, ch chan<- signResult) {
	fid := c.getFid(kw)
	if fid == "-1" {
		ch <- signResult{kw: kw, resp: ""}
		return
	}
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
	r, err := http.NewRequest("POST", signUrl, strings.NewReader(formData.Encode()))
	if err != nil {
		ch <- signResult{kw: kw, resp: ""}
		return
	}
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("User-Agent", "Pkmm iPhone/1.0 BadApple/99.1")
	r.AddCookie(c.cookie)
	resp, err := c.client.Do(r)
	if err != nil {
		ch <- signResult{kw: kw, resp: ""}
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		ch <- signResult{kw: kw, resp: ""}
		return
	}
	ch <- signResult{kw: kw, resp: string(body)}
}

// 签到所有贴吧，使用goroutines
func (c *Crawl) SignAll(tiebas []string) *map[string]tiebaSignInfo {
	goroutineCount := 20 // 最多开启20个线程
	needSignTiebaChans := make(chan string, goroutineCount)

	go func() {
		for _, tieba := range tiebas {
			needSignTiebaChans <- tieba
		}
		close(needSignTiebaChans)
	}()

	resultChans := make(chan signResult, goroutineCount)
	go func() {
		for {
			tieba, ok := <-needSignTiebaChans
			if !ok {
				break
			}
			go c.signOne(tieba, resultChans)
		}
	}()
	size := len(tiebas)
	result := make(map[string]tiebaSignInfo, size)
	for i := 0; i < size; i++ {
		ret := <-resultChans
		t := tiebaSignInfo{}
		json.Unmarshal([]byte(ret.resp), &t)
		result[ret.kw] = t
	}
	return &result
}

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

// 定时在每天凌晨执行(北京时间)
func (c *Crawl) RunAtDaily() {
	go func() {
		for {
			now := time.Now()
			next := now.Add(time.Hour * 24)
			next = time.Date(next.Year(), next.Month(), next.Day(), 0, 0, 0, 0, next.Location())
			t := time.NewTicker(next.Sub(now))
			<-t.C
			retry := 3
			for retry > 0 {
				retry--
				tiebas, err := c.RetrieveTiebas()
				if err != nil {
					//todo log
					continue
				}
				c.SignAll(tiebas)
				break
			}
		}
	}()
}
