package linedata

import (
	"ad-data/function"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

var (
	//全局集合，用于存储line粉丝数
	fansMap map[string]int
	//全局通道，用于存储line信息
	//lineChan chan string
	//全局通道，用于存储line@粉丝数
	fansChan chan int
	//并发控制器，打组
	wg sync.WaitGroup
)

//Line 业务号结构,主要用于获取line的信息
type Line struct {
	AllAccount map[string]string
	AllFans map[string]int
}

//TotalFans 所有line@账号的粉丝数,最主要也是最常用的一个方法
func (line *Line) TotalFans() int {
	var sumFans int
	fansChan = make(chan int)
	defer close(fansChan)
	go func() {
		for {
			select {
			case n := <-fansChan:
				sumFans += n
			default:

			}
		}
	}()

	wg.Add(len(line.AllAccount))
	for _, v := range line.AllAccount {
		go line.Fans(v)
	}
	wg.Wait()

	//time.Sleep(3 *time.Second)

	if sumFans == 0 {
		fmt.Println("查询超时！！")
	}

	return sumFans
}

//Fans 获取line@账号的粉丝数，不返回，传入通道
func (line *Line) Fans(account string) {
	defer wg.Done()
	account = url.QueryEscape(account)
	var fans int
	webURL := "https://social-plugins.line.me/widget/friend?lineId=" + account + "&count=true&lang=en&origin=https://www." + function.RandString(5) + ".html&title=" + function.RandString(5)
	res, err := http.Get(webURL)
	if err != nil || res.StatusCode != 200 {
		log.Fatal("无法访问：", webURL)
	}
	defer res.Body.Close()

	//加载HTML
	doc, e := goquery.NewDocumentFromReader(res.Body)
	if e != nil {
		log.Fatal(e)
	}
	doc.Find(".btnWrap").Find(".bubble").Find(".num").Each(
		func(i int, s *goquery.Selection) {
			res, ok := s.Attr("title")
			if !ok {
				log.Fatal("查找不到粉丝数量！")
			}
			fans, err = strconv.Atoi(res)
			if err != nil {
				log.Fatal("字符串类型转化错误")
			}
		})

	fansChan <- fans
}

//GetFans @账号的粉丝数，并返回
func GetFans(account string) (fans int) {
	account = url.QueryEscape(account)
	webURL := "https://social-plugins.line.me/widget/friend?lineId=" + account + "&count=true&lang=en&origin=https://www." + function.RandString(5) + ".html&title=" + function.RandString(5)
	res, err := http.Get(webURL)
	if err != nil || res.StatusCode != 200 {
		log.Fatal("无法访问：", webURL)
	}
	defer res.Body.Close()

	//加载HTML
	doc, e := goquery.NewDocumentFromReader(res.Body)
	if e != nil {
		log.Fatal(e)
	}
	doc.Find(".btnWrap").Find(".bubble").Find(".num").Each(
		func(i int, s *goquery.Selection) {
			res, ok := s.Attr("title")
			if !ok {
				log.Fatal("查找不到粉丝数量！")
			}
			fans, err = strconv.Atoi(res)
			if err != nil {
				log.Fatal("字符串类型转化错误")
			}
		})

	return
}

//GetAll 获取所有的信息
func (line *Line) GetAll() {
	fansMap = make(map[string]int, len(line.AllAccount))
	for n, v := range line.AllAccount {
		line.GroupString(v, n)
	}
}

//GroupString 查询【业务员】的【line账号】信息
func (line *Line) GroupString(v string, n string) {
	var name string
	if v != "" {
		s := n + v
		account := url.QueryEscape(v)
		fansMap[s], name = line.GetLineData(account)
		var build strings.Builder
		build.WriteString(s)
		build.WriteString(" 粉丝数：")
		build.WriteString(strconv.Itoa(fansMap[s]))
		build.WriteString(" 昵称：")
		build.WriteString(name)
		//lineChan <- build.String()
		fmt.Println(build.String())
	} else {
		fmt.Println("账号为空！")
	}
}

//GetLineData 获取line信息(包含粉丝数和昵称）
func (line *Line) GetLineData(account string) (fans int, name string) {
	//请求HTMLhttps://social-plugins.line.me/widget/friend?lineId=%40dtd3861q&count=true&lang=en&origin=https%3A%2F%2Fsocial-plugins.line.me%2Fen%2Fhow_to_install&title=LINE%20Social%20Plugins
	webURL := "https://social-plugins.line.me/widget/friend?lineId=" + account + "&count=true&lang=en&origin=https://www." + function.RandString(5) + ".html&title=" + function.RandString(5)
	fmt.Println(webURL)
	res, err := http.Get(webURL)
	if err != nil || res.StatusCode != 200 {
		log.Fatal("无法访问：", webURL)
	}
	defer res.Body.Close()

	//加载HTML
	doc, e := goquery.NewDocumentFromReader(res.Body)
	if e != nil {
		log.Fatal(e)
	}
	doc.Find(".btnWrap").Find(".bubble").Find(".num").Each(
		func(i int, s *goquery.Selection) {
			res, ok := s.Attr("title")
			if !ok {
				log.Fatal("查找不到粉丝数量！")
			}
			fans, err = strconv.Atoi(res)
			if err != nil {
				log.Fatal("字符串类型转化错误")
			}
		})

	doc.Find(".btnWrap").Find(".btn").Each(
		func(i int, s *goquery.Selection) {
			res, exists := s.Attr("title")
			if !exists {
				log.Fatal("查找不到账号昵称")
			}
			name = line.getName(res)
		})

	return fans, name
}

//getName 用正则表达式提取账号昵称
//需要从字符串内提取出：資深護膚老師-Amy
func (line *Line) getName(str string) string {
	as := strings.Index(str, "as") - 1
	// return strings.Trim(str[4:as], " ")
	return str[4:as]
}
