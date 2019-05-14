package main

import (
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	"strings"
	"github.com/robfig/cron"
)

var (
	APPID          = "*********"
	APPSECRET      = "**************"
	SentTemplateID = "0U-LzrPOkwGsg0GwCEcNuEfmotI-JYsX60WqO0tIJdo"	//每日一句的模板ID，替换成自己的
	WeatTemplateID = "YBF_WVK2Bbq2ChkDa4PGecZ8UnQsqugqbfolZedP-Sc"  //天气模板ID，替换成自己的
	WeatherVersion = "v1"
)

type token struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type sentence struct {
	Content     string `json:"content"`
	Note        string `json:"note"`
	Translation string `json:"translation"`
}

func main() {
	spec := "0 0 12 * * *" // 每天12:00
	spec1 := "0 0 7 * * *" // 每天早晨7:00
	c := cron.New()
	c.AddFunc(spec, everydaysen)
	c.AddFunc(spec1, weather)
	c.Start()
	fmt.Println("开启定时任务")
	select {}
	//weather()
	//everydaysen()
}

//发送每日一句
func everydaysen() {
	req, fxurl := getsen()
	if req.Content == "" {
		return
	}
	access_token := getaccesstoken()
	if access_token == "" {
		return
	}

	flist := getflist(access_token)
	if flist == nil {
		return
	}

	reqdata := "{\"content\":{\"value\":\"" + req.Content + "\", \"color\":\"#0000CD\"}, \"note\":{\"value\":\"" + req.Note + "\"}, \"translation\":{\"value\":\"" + req.Translation + "\"}}"
	for _, v := range flist {
		templatepost(access_token, reqdata, fxurl, SentTemplateID, v.Str)
	}
}

//发送天气预报
func weather() {
	access_token := getaccesstoken()
	if access_token == "" {
		return
	}

	flist := getflist(access_token)
	if flist == nil {
		return
	}

	var city string
	for _, v := range flist {
		switch v.Str {
		case "oeZ6P5kyGsLKn3sIGRVfpb8oT4mg":
			city = "青岛"
			go sendweather(access_token, city, v.Str)
		case "oeZ6P5jvFNh2y_h_2UcaoTXBaC2o":
			city = "西安"
			go sendweather(access_token, city, v.Str)
		default:

		}
	}
	fmt.Println("weather is ok")
}

//获取微信accesstoken
func getaccesstoken() string {
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%v&secret=%v", APPID, APPSECRET)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("获取微信token失败", err)
		return ""
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("微信token读取失败", err)
		return ""
	}

	token := token{}
	err = json.Unmarshal(body, &token)
	if err != nil {
		fmt.Println("微信token解析json失败", err)
		return ""
	}

	return token.AccessToken
}

//获取每日一句
func getsen() (sentence, string) {
	resp, err := http.Get("http://open.iciba.com/dsapi/?date")
	sent := sentence{}
	if err != nil {
		fmt.Println("获取每日一句失败", err)
		return sent, ""
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("读取内容失败", err)
		return sent, ""
	}

	err = json.Unmarshal(body, &sent)
	if err != nil {
		fmt.Println("每日一句解析json失败")
		return sent, ""
	}
	fenxiangurl := gjson.Get(string(body), "fenxiang_img").String()
	fmt.Println(sent)
	return sent, fenxiangurl
}

//获取关注者列表
func getflist(access_token string) []gjson.Result {
	url := "https://api.weixin.qq.com/cgi-bin/user/get?access_token=" + access_token + "&next_openid="
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("获取关注列表失败", err)
		return nil
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("读取内容失败", err)
		return nil
	}
	flist := gjson.Get(string(body), "data.openid").Array()
	return flist
}

//发送模板消息
func templatepost(access_token string, reqdata string, fxurl string, templateid string, openid string) {
	url := "https://api.weixin.qq.com/cgi-bin/message/template/send?access_token=" + access_token

	reqbody := "{\"touser\":\"" + openid + "\", \"template_id\":\"" + templateid + "\", \"url\":\"" + fxurl + "\", \"data\": " + reqdata + "}"

	resp, err := http.Post(url,
		"application/x-www-form-urlencoded",
		strings.NewReader(string(reqbody)))
	if err != nil {
		fmt.Println(err)
		return
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(body))
}

//获取天气
func getweather(city string) (string, string, string, string) {
	url := fmt.Sprintf("https://www.tianqiapi.com/api?version=%s&city=%s", WeatherVersion, city)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("获取天气失败", err)
		return "", "", "", ""
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("读取内容失败", err)
		return "", "", "", ""
	}

	data := gjson.Get(string(body), "data").Array()
	thisday := data[0].String()
	day := gjson.Get(thisday, "day").Str
	wea := gjson.Get(thisday, "wea").Str
	tem := gjson.Get(thisday, "tem").Str
	//tem2 := gjson.Get(thisday, "tem2").Str
	air_tips := gjson.Get(thisday, "air_tips").Str
	return day, wea, tem, air_tips
}

//发送天气
func sendweather(access_token, city, openid string) {
	day, wea, tem, air_tips := getweather(city)
	if day == "" || wea == "" || tem == ""|| air_tips == "" {
		return
	}
	reqdata := "{\"city\":{\"value\":\"城市：" + city + "\", \"color\":\"#0000CD\"}, \"day\":{\"value\":\"" + day + "\"}, \"wea\":{\"value\":\"天气：" + wea + "\"}, \"tem1\":{\"value\":\"平均温度：" + tem + "\"}, \"air_tips\":{\"value\":\"tips：" + air_tips + "\"}}"
	//fmt.Println(reqdata)
	templatepost(access_token, reqdata, "http://47.105.41.130:9999", WeatTemplateID, openid)
}
