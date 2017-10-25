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
	TBS_URL    = "http://tieba.baidu.com/dc/common/tbs"
	FID_URL    = "http://tieba.baidu.com/mo/m"
	I_LIKE_URL = "http://tieba.baidu.com/mo/q-0-" +
		"-49BB6FC27D7013BD795602B74B2E83E2%3" +
		"AFG%3D1--1-1-0----wapp_1462281637540_9" +
		"23/m?tn=bdFBW&tab=favorite"
	SIGN_URL = "http://c.tieba.baidu.com/c/c/forum/sign"
)

var client = &http.Client{}

func GetTbs(bduss string) string {
	cookie := http.Cookie{Name: "BDUSS", Value: bduss}
	req, _ := http.NewRequest("GET", TBS_URL, nil)
	req.AddCookie(&cookie)
	rep, err := client.Do(req)
	data, _ := ioutil.ReadAll(rep.Body)
	defer rep.Body.Close()
	type TBS struct {
		Tbs string
	}
	var tbs TBS
	if err == nil {
		json.Unmarshal(data, &tbs)
		return tbs.Tbs
	}
	return ""
}

func GetFid(kw, bduss string) string {
	cookie := http.Cookie{Name: "BDUSS", Value: bduss}
	args := url.Values{}
	args.Add("kw", kw)
	request, _ := http.NewRequest("GET", FID_URL+"?"+args.Encode(), nil)
	request.AddCookie(&cookie)
	rep, err := client.Do(request)
	if err != nil {
		return "-1"
	}
	defer rep.Body.Close()
	data, _ := ioutil.ReadAll(rep.Body)
	re, _ := regexp.Compile(`<input type="hidden" name="fid" value="(.*?)"/>`)
	submatch := re.FindSubmatch(data)
	if len(submatch) < 2 {
		return "-1"
	}
	return string(submatch[1])
}

func GetAllstar(bduss string) []string {
	cookie := http.Cookie{Name: "BDUSS", Value: bduss}
	req, _ := http.NewRequest("GET", I_LIKE_URL, nil)
	req.AddCookie(&cookie)
	rep, err := client.Do(req)
	if err != nil {
	}
	defer rep.Body.Close()
	data, _ := ioutil.ReadAll(rep.Body)
	re, _ := regexp.Compile(`<a href=".*?kw=.*?">(.*?)</a>`)
	submatch := re.FindAllSubmatch(data, -1)
	ret := []string{}
	for _, name := range submatch {
		if len(name) < 2 {
			continue
		}
		ret = append(ret, string(name[1]))
	}
	return ret
}

func Sign(kw, fid, bduss string, ch chan string) {
	cookie := http.Cookie{Name: "BDUSS", Value: bduss}
	data := url.Values{
		"BDUSS":           {bduss},
		"_client_id":      {"03-00-DA-59-05-00-72-96-06-00-01-00-04-00-4C-43-01-00-34-F4-02-00-BC-25-09-00-4E-36"},
		"_client_type":    {"4"},
		"_client_version": {"1.2.1.17"},
		"_phone_imei":     {"540b43b59d21b7a4824e1fd31b08e9a6"},
		"fid":             {fid},
		"kw":              {kw},
		"net_type":        {"3"},
		"tbs":             {GetTbs(bduss)},
	}
	var keys = make([]string, len(data))
	var i = 0
	for k, _ := range data {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	var enerypt bytes.Buffer
	for _, k := range keys {
		enerypt.WriteString(k)
		enerypt.WriteString("=")
		enerypt.WriteString(data.Get(k))
	}
	enerypt.WriteString("tiebaclient!!!")
	tt := md5.New()
	tt.Write(enerypt.Bytes())
	sign := hex.EncodeToString(tt.Sum(nil))
	sign = strings.ToUpper(sign)
	data.Set("sign", sign)
	req, _ := http.NewRequest("POST", SIGN_URL, strings.NewReader(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("User-Agent", "Pkmm iPhone/1.0 BadApple/99.1")
	req.AddCookie(&cookie)
	rep, _ := client.Do(req)
	defer rep.Body.Close()
	html, _ := ioutil.ReadAll(rep.Body)
	fmt.Println(string(html))
	type ResponseJson struct {
		Error_code string
		Error_msg  string
		Info       []string
		Error      map[string]string
	}
	ret := ResponseJson{}
	json.Unmarshal(html, &ret)
	ch <- string(html)
	ch <- kw
}
