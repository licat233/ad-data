package addata

import (
	"ad-data/project"
	"fmt"
	"time"
)

func InsertProject(PJ project.Project,) error {
	return PJ.NewProject()
}

type Param struct {
	Projectname string `form:"projectname" json:"projectname" xml:"projectname" binding:"required"`
	AccountID string `form:"AccountID" json:"AccountID" xml:"AccountID" binding:"required"`
	CampaignId string `form:"CampaignId" json:"CampaignId" xml:"CampaignId" binding:"required"`
	Linename string `form:"linename" json:"linename" xml:"linename" binding:"required"`
	Lineid string `form:"lineid" json:"lineid" xml:"lineid" binding:"required"`
}

var TimingChan = make(chan bool)
var (
	//更新时间间隔,单位秒
	t time.Duration = 10
	TimingStatus bool
)
//var moneymsg = make(chan error)
//var fansmsg = make(chan error)
//pJ是主要存储项目信息的结构体

//TimingMoney 每隔t秒更新一次所有活动的消耗
func TimingMoney(PJ project.Project,t time.Duration) {
	fmt.Println("开始监控消耗数据")
	for {
		select {
		case <-TimingChan:
			fmt.Println("已关闭消耗数据监控")
			TimingStatus = false
			break
			return
		default:
		}
		now := time.Now().Format("2006-01-02 15:04")
		for _, c := range PJ.Campaignid {
			go func() {
				if err := project.Mgopopin.UpdatePopin(c); err != nil {
					fmt.Printf("%s粉丝更新失败%s\n", c, now)
				}
			}()
		}
		<-time.NewTimer(time.Second * t).C
	}
}

//TimingFans 每隔t秒更新一次所有line的粉丝数
func TimingFans(PJ project.Project,t time.Duration) {
	fmt.Println("开始监控加粉数据")
	for {
		select {
		case <-TimingChan:
			fmt.Println("已关闭加粉数据监控")
			TimingStatus = false
			break
			return
		default:
		}
		now := time.Now().Format("2006-01-02 15:04")
		for _, a := range PJ.Lineid {
			go func() {
				if err := project.Mgoline.UpdateLine(a); err != nil {
					fmt.Printf("%s粉丝更新失败%s\n", a, now)
				}
			}()
		}
		<-time.NewTimer(time.Second * t).C
	}
}

//TimingProject 每隔t秒更新一次开启状态的广告项目信息
func TimingProject(PJ project.Project,t time.Duration) {
	if PJ.Status != 0 {
		go TimingMoney(PJ,t)
		go TimingFans(PJ,t)
	} else {
		fmt.Printf("%s项目已关闭！请开启\n", PJ.Name)
	}
	<-time.NewTimer(time.Second * (t + 60)).C
	fmt.Printf("开始更新%s项目信息\n", PJ.Name)
	TimingStatus = true
	for {
		select {
		case <-TimingChan:
			fmt.Println("已关闭项目更新")
			TimingStatus = false
			break
			return
		default:
		}
		if err:=PJ.UpdateProject();err != nil{
			fmt.Printf("%s项目更新失败%s\n", PJ.Name, time.Now().Format("2006-01-02 15:04"))
		}
		<-time.NewTimer(time.Second * t).C
	}
}

func Projectdata(pid string) interface{} {
	res,_:=project.GetProjectidData(pid)
	//fmt.Println(res)
	return res
}

func Todaydata(pid string) (interface{}, error){
	res,err:=project.GetProjectDataNewest(pid)
	if err != nil{
		return res,err
	}
	return res,err
}

func APdatabyid(pid string) interface{} {
	res,_:=project.GetProDataByid(pid)
	return res
}

func GetAllProjectData() interface{} {
	res,_:=project.GetAllProjectData()
	return res
}

func AddProject(param Param) error {
	//实例化项目结构
	P := &project.Project{
		Name:       param.Projectname,
		Account:     param.AccountID,
		Campaignid: []string{param.CampaignId},
		Lineid:     map[string]string{ param.Linename : param.Lineid },
		Status:     0,
	}
	return P.NewProject()
}

func MdfProjectStatus(pid string,s int) error {
	return project.MdfProjectStatus(pid,s)
}

func InstructTiming(i string) {
	if i == "on" && TimingStatus == false {
		PJ := project.OnProject()
		go TimingProject(PJ,t)
	}else if i == "off" {
		TimingChan<-true
	}
}