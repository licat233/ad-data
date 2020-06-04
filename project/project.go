package project

import (
	"ad-data/config"
	"ad-data/function"
	"ad-data/linedata"
	"ad-data/mongodb"
	"ad-data/popindata"
	"errors"
	"fmt"
	"log"

	"gopkg.in/mgo.v2/bson"
)

const (
	database   string = "Project"
	collection string = "project"
)

//Project 项目结构
type Project struct {
	//Id         bson.ObjectId     `bson:"_id"`
	Name       string              `bson:"name"`
	Campaignid []string            `bson:"campaignID"`
	Lineid     map[string]string   `bson:"lineid"`
	Account    string              `bson:"account"`
	Status     int                 `bson:"status"`
	Date       bson.MongoTimestamp `bson:"date"`
}

//ChildProject 子项目结构
type ChildProject struct {
	//Id         bson.ObjectId `bson:"_id"`
	Name       string              `bson:"name"`
	Totalmoney int                 `bson:"totalmoney"`
	Spend      int                 `bson:"spend"`
	ToalFans   int                 `bson:"totalfans"`
	Fans       int                 `bson:"fans"`
	Cpa        float32             `bson:"cpa"`
	Date       bson.MongoTimestamp `bson:"date"`
}

var (
	//实例化项目结构、、核心配置
	PJ       = OnProject()
	Mgopopin = &popindata.MgoPopin{
		Account:  config.POPaccount,
		Cookie: config.POPCookie,
			Campaign: PJ.Campaignid,
	}

	Mgoline = &linedata.MgoLine{}
)

//从数据库获取所需要监控的项目，status == 1
func OnProject() Project {
	var Pstruct Project
	err := mongodb.FindA(database, collection, bson.M{"status": 1}, &Pstruct)
	if err != nil {
		err = errors.New("没有开启状态的项目")
	}
	return Pstruct
}

//NewProject 新建项目
func (project *Project) NewProject() (err error) {
	p := Project{
		Name:       project.Name,
		Campaignid: project.Campaignid,
		Lineid:     project.Lineid,
		Account:    project.Account,
		Status:     project.Status,
	}
	//注册项目
	if err = mongodb.Insert(database, collection, &p); err != nil {
		log.Fatalf("%s 项目注册失败: %s\n", p.Name, err)
	}
	//注册或更新line
	Mgoline.RegisterLine(p.Lineid)
	//注册或更新POPin活动
	Mgopopin.RegisterPopin(p.Campaignid)
	//新增ChildProject项目
	//创建一个项目名的集合，文档为各时间段的信息
	totalmoney := Mgopopin.FindAllMoney(p.Campaignid) //时效性：即时数据
	totalfans := Mgoline.FindAllFans(p.Lineid)        //时效性：即时数据
	childproject := &ChildProject{
		Name:       p.Name,
		Totalmoney: totalmoney,
		Spend:      0,
		ToalFans:   totalfans,
		Fans:       0,
		Cpa:        0.00,
	}

	//注册单个文档记录,集合不存在会自动创建
	pid, err := mongodb.FindID(database, collection, bson.M{"name": p.Name})
	if err != nil {
		fmt.Println(project.Name, "项目ID查询失败", err)
	}
	fmt.Println(childproject)
	if err = mongodb.Insert(database, pid, &childproject); err != nil {
		fmt.Println(project.Name, "子项目注册失败", err)
	}
	return
}

//更新子项目信息
func (project *Project) UpdateProject() error {
	pid, err := mongodb.FindID(database, collection, bson.M{"name": project.Name})
	if err != nil {
		fmt.Println(project.Name, "项目ID查询失败", err)
	}
	totalmoney := project.GetlastMoney()
	totalfans := project.GetlastFans()
	Spend := totalmoney - project.GetMarkMoney()
	Fans := totalfans - project.GetMarkFans()
	var cpa float32
	if Fans == 0 {
		cpa = float32(Spend / 1)
	} else {
		cpa = float32(Spend / Fans)
	}
	childproject := &ChildProject{
		Name:       project.Name,
		Totalmoney: totalmoney,
		Spend:      Spend,
		ToalFans:   totalfans,
		Fans:       Fans,
		Cpa:        cpa,
	}
	err = mongodb.Insert(database, pid, &childproject)
	if err != nil {
		fmt.Println(project.Name, "子项目更新失败", err)
	}
	return err
}

//获取某些活动的最新总消耗BY数据库
func (project *Project) GetlastMoney() int {
	return Mgopopin.GetNewestMoney(project.Campaignid)
}

//获取某些活动的最新粉丝数BY数据库
func (project *Project) GetlastFans() int {
	return Mgoline.GetNewestFans(project.Lineid)
}

//获取某些活动的Mark总消耗BY数据库
func (project *Project) GetMarkMoney() int {
	return Mgopopin.GetMarkMoney(project.Campaignid)
}

//获取某些活动的Mark粉丝数BY数据库
func (project *Project) GetMarkFans() int {
	return Mgoline.GetMarkFans(project.Lineid)
}

//获取某个Campaign的数据
func GetProjectidData(pid string) (interface{}, error) {
	//Chart图标数据结构
	type Chart struct {
		Date int64 `bson:"date"`
		CNY  int   `bson:"spend"`
		CV   int   `bson:"fans"`
		CPA  int   `bson:"cpa"`
	}
	var reschart []Chart
	err := mongodb.FindAllData(database, pid, &reschart)
	for i := 0; i < len(reschart); i++ {
		reschart[i].Date = function.GetUnixTime(reschart[i].Date)
	}
	return reschart, err
}

//获取某个项目的最新消耗加粉数据
func GetProjectDataNewest(pid string) (interface{}, error) {
	type TodayStruct struct {
		//Id         bson.ObjectId `bson:"_id"`
		Name       string  `bson:"name"`
		Totalmoney int     `bson:"totalmoney"`
		Spend      int     `bson:"spend"`
		ToalFans   int     `bson:"totalfans"`
		Fans       int     `bson:"fans"`
		Cpa        float32 `bson:"cpa"`
		Date       int64   `bson:"date"`
	}
	var resdata []TodayStruct
	//fmt.Println("MARKE:",pid)
	err := mongodb.FindNewest(database, pid, &resdata)
	if len(resdata) == 0 {
		return nil, errors.New("此项目无数据")
	}
	resdata[0].Date = function.GetUnixTime(resdata[0].Date)
	return resdata[0], err
}

//GetProDataByid 获取某个项目的活动信息
func GetProDataByid(pid string) (interface{}, error) {
	var resdata Project
	id := bson.ObjectIdHex(pid)
	err := mongodb.FindA(database, collection, bson.M{"_id": id}, &resdata)
	return resdata, err
}

func GetAllProjectData() (interface{}, error) {
	type Projects struct {
		PId        bson.ObjectId       `bson:"_id"`
		Name       string              `bson:"name"`
		Campaignid []string            `bson:"campaignID"`
		Lineid     map[string]string   `bson:"lineid"`
		Account    string              `bson:"account"`
		Status     int                 `bson:"status"`
		Date       bson.MongoTimestamp `bson:"date"`
	}
	var resdata []Projects
	err := mongodb.FindAll(database, collection, nil, nil, &resdata)
	return resdata, err
}

func MdfProjectStatus(pid string, s int) error {
	id := bson.ObjectIdHex(pid)
	return mongodb.Update(database, collection, bson.M{"_id": id}, bson.M{"$set": bson.M{"status": s}})
}
