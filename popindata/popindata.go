package popindata

import (
	"ad-data/function"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

var (
	wg sync.WaitGroup
)

type popinJSONObj struct {
	TotalMoney string `json:"total_charge"`
}

//GetAllMoney 获取popin指定campaign的总消耗,最常用的
func (mgoPopin *MgoPopin) GetAllMoney() (money int) {
	for i, v := range mgoPopin.Campaign {
		j := mgoPopin.GetMoney(mgoPopin.Cookie, mgoPopin.Account, v)
		fmt.Println(i, ":", j)
		money += j
	}
	return
}

//AllOnMoney 获取popin的ON活动的总消耗,很少用
func (mgoPopin *MgoPopin) AllOnMoney() (money int) {
	idRes := mgoPopin.GetCampaignID(mgoPopin.Cookie, mgoPopin.Account)
	for i, v := range idRes {
		j := mgoPopin.GetMoney(mgoPopin.Cookie, mgoPopin.Account, v)
		fmt.Println(i, ":", j)
		money += j
	}
	return
}

//GetMoney 爬取所需要的数据（消耗额）,核心函数,不要手贱的去"优化"
func (mgoPopin *MgoPopin) GetMoney(cookie, accountID, campaignID string) int {
	newcookie := cookie
    RECOOKIE:
	url := "https://dashboard.popin.cc/discovery/accounts-tw/index.php/" + accountID + "/c/" + campaignID + "/getDiscoveryReports"
	referer := "https://dashboard.popin.cc/discovery/accounts-tw/index.php/" + accountID + "/manageAgency?subPage=/" + accountID + "/campaigns/listCampaign"

	client := &http.Client{}
	var resp *http.Response
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("http.NewRequest Failed:", err)
		//os.Exit(1)
	}

	req.Header.Add("Host", "dashboard.popin.cc")
	req.Header.Add("User-Agent", function.GetRandomUserAgent())
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Accept-Language", "zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2")
	req.Header.Add("Accept-Encoding", "gzip, deflate, br")
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Referer", referer)
	req.Header.Add("Cookie", newcookie)
	req.Header.Add("Pragma", "no-cache")
	req.Header.Add("Cache-Control", "no-cache")

	//发送请求数据之服务器，并获取响应数据
	resp, err = client.Do(req)
	//time.Sleep(5 * time.Second)

	if err != nil {
		log.Fatal("Client Do Failed:", err)
		//os.Exit(1)
	}
	if resp.StatusCode != 200 {
		log.Fatal("status code error: ", resp.StatusCode, resp.Status)
		//os.Exit(1)
	}

	// 打印所有请求头信息
	//for k, v := range req.Header {
	//	fmt.Println("请求头信息 :", k, "=", v)
	//}

	// 打印所有响应头信息
	//for k, v := range resp.Header {
	//	fmt.Println("响应头信息 :", k, "=", v)
	//}

	//读取JSON文件内容 返回字节切片
	bytesSlice, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		log.Fatal("ioutil.ReadAll failed:", err)
		//os.Exit(1)
	}

	fmt.Println(newcookie)
	//开始处理信息
	var popJSON popinJSONObj
	//解码json数据，将字节切片映射到指定结构上
	e := json.Unmarshal(bytesSlice, &popJSON)
	if e != nil {
		//表示 cookie已经过期啦
		//fmt.Println("cookie不可用！json.Unmarshal:", e)
		log.Fatal("cookie不可用！json.Unmarshal:", e)
		newcookie = GetPopinCookie()
		goto RECOOKIE
		//os.Exit(1)
	}

	//fmt.Println(popJSON.TotalMoney)

	moneystring := strings.Replace(popJSON.TotalMoney, ",", "", -1)
	moneystring = moneystring[:len(moneystring)-3]

	moneyint, err := strconv.Atoi(moneystring)
	if err != nil {
		log.Fatal("strconv.Atoi failed:", err)
		//os.Exit(1)
	}

	return moneyint
}

//GetCampaignID 获取ON活动的ID
func (mgoPopin *MgoPopin) GetCampaignID(cookie, accountID string) []string {
	url := "https://dashboard.popin.cc/discovery/accounts-tw/index.php/" + accountID + "/campaigns/listCampaign"
	referer := "https://dashboard.popin.cc/discovery/accounts-tw/index.php/" + accountID + "/manageAgency?subPage=/" + accountID + "/campaigns/listCampaign"
	client := &http.Client{}
	var resp *http.Response
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal("http.NewRequest Failed:", err)
		//os.Exit(1)
	}

	req.Header.Add("Host", "dashboard.popin.cc")
	req.Header.Add("User-Agent", function.GetRandomUserAgent())
	req.Header.Add("Accept", "text/html, */*; q=0.01")
	req.Header.Add("Accept-Language", "zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2")
	req.Header.Add("Accept-Encoding", "gzip, deflate, br")
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Referer", referer)
	req.Header.Add("Cookie", cookie)
	req.Header.Add("Pragma", "no-")
	req.Header.Add("Cache-Control", "no-cache")

	//发送请求数据之服务器，并获取响应数据
	resp, err = client.Do(req)
	if err != nil {
		log.Fatal("Client Do Failed:", err)
		//os.Exit(1)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Fatal("status code error:", resp.StatusCode, resp.Status)
		//os.Exit(1)
	}

	byt, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("ioutil.ReadAll failed:", err)
		//os.Exit(1)
	}
	content := string(byt)
	//fmt.Println(content)

	//正则表达式
	regID := regexp.MustCompile(`data-campaignid="[a-zA-Z0-9]+"`)
	regStatus := regexp.MustCompile(`campaign-status\sstring">[A-Z]+</td>`)
	idSlice := regID.FindAllString(content, -1)
	statusSlice := regStatus.FindAllString(content, -1)
	//var statusChan chan int
	statusChan := make(chan int)
	defer close(statusChan)
	var campaignID []string

	go func() {
		for {
			select {
			case i := <-statusChan:
				v := idSlice[i][17 : len(idSlice[i])-1] //已获取ON的活动ID
				campaignID = append(campaignID, v)
			default:

			}
		}
	}()

	for i, va := range statusSlice {
		if strings.Index(va, "ON") != -1 {
			//va = va[24 : len(va)-5]
			statusChan <- i
		}
	}

	return campaignID
}

func GetPopinCookie()(cookie string) {
	requrl := "https://dashboard.popin.cc/discovery/accounts-tw/index.php"
	data := make(url.Values)
	data["userid"] = []string{"Self_White"}
	data["password"] = []string{"Self_White_..."}
	data["autoLoginEnable"] = []string{"on"}
	res, err := http.PostForm(requrl, data)
	//设置http中header参数，可以再此添加cookie等值
	res.Header.Add("Host", "dashboard.popin.cc")
	res.Header.Add("User-Agent", function.GetRandomUserAgent())
	//res.Header.Add("User-Agent", "***")
	//res.Header.Add("http.socket.timeou", 5000)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer res.Body.Close()
	cookieSlices := res.Header["Set-Cookie"]
	PHPSESSID := cookieSlices[0]
	laravel_session := cookieSlices[1]
	reg1 := regexp.MustCompile("PHPSESSID=\\S*")
	reg2 := regexp.MustCompile("laravel_session=\\S*")
	cookie1 := reg1.FindAllString(PHPSESSID, -1)
	cookie2 := reg2.FindAllString(laravel_session, -1)
	var buffer bytes.Buffer
	buffer.WriteString(cookie1[0])
	buffer.WriteString(" ")
	buffer.WriteString(cookie2[0])
	buffer.WriteString(" __stripe_mid=abbaa6ef-6d4b-44f7-82b7-6f3d7da3ca68; autoLoginEnable=true; ticket=94c85e7803; __zlcmid=xdj2wiKdOzUtvl")
	cookie = buffer.String()
	//if strings.TrimSpace(cookie);cookie==""{
	//	return errors.New("cookie not found")
	//}
	return
}