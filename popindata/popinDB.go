package popindata

import (
	"ad-data/function"
	"ad-data/mongodb"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"reflect"
	"sync"
	"time"
)

const (
	database   = "Popin"
	collection = "campaign"
)

//popindata重点在于更新消耗tatolmoney
//MgoPopin Mgo活动结构,主要存储popin账户信息
type MgoPopin struct {
	//Id      bson.ObjectId `bson:"_id"`
	Account  string
	Cookie   string
	Campaign []string
	//AllMoney int
}

//MgoPopin 活动结构，主要存储popin活动信息
type MyPopin struct {
	//Id         bson.ObjectId `bson:"_id"`
	CampaignID string        `bson:"campaignID"`
	TotalMoney int           `bson:"totalmoney"`
	Status     int           `bson:"status"`
	Date        bson.MongoTimestamp         `bson:"date"`
}

var (
	dbMutex sync.RWMutex
	//全局通道，用于存储总金额
	ttmoney chan int
)
//操作时注意读写锁的使用，防止死锁、活锁、饥饿锁的发生
//主要逻辑：
//注册：存在则更新
//更新：更新之后在指定集合插入新的文档，集合不存在则创建
//注册
func (mgoPopin *MgoPopin) RegisterPopin(allCampaign []string) {
	wg.Add(len(allCampaign))
	for _, c := range allCampaign {
		go mgoPopin.mgoInsert(c)
	}
	wg.Wait()
}

func (mgoPopin *MgoPopin) mgoInsert(c string) {
	defer wg.Done()
	money := mgoPopin.GetMoney(mgoPopin.Cookie, mgoPopin.Account, c)
	//数据库加锁
	//dbMutex.Lock()
	p := MyPopin{
		CampaignID: c,
		TotalMoney: money,
		Status:     0,
		//Date:       now,
	}
	//更新，如果不存在就插入一个新的数据
	if err:= mongodb.Upsert(database,collection,bson.M{"campaignID": c},&p);err != nil {
		fmt.Println(c, "POPin项目更新注册失败",err)
	}
	////////旧方案
	//err := mongodb.FindA(database, collection, bson.M{"campaignID": c}, &p)
	//if err == nil {
	//	//说明已经存在，更新即可
	//	err = mongodb.Update(database, collection, bson.M{"campaignID": c}, bson.M{"$set": bson.M{"totalmoney": money, "date": now}})
	//	if err != nil {
	//		fmt.Println(c, "更新失败",err)
	//	}
	//	return
	//}
	//
	//if err = mongodb.Insert(database, collection, &p); err != nil {
	//	fmt.Println(c, "注册失败",err)
	//}
	////////////////////////

	//注册单个line account文档对应的集合，不存在会自动创建,以插入的方式更新
	if err := mongodb.Insert(database, c, &p); err != nil {
		fmt.Println(c, "POPin子项目注册失败",err)
	}

	//数据库解锁
	//dbMutex.Unlock()
	return
}

//FindPopin 查询popin活动的信息
func (mgoPopin *MgoPopin) FindPoin(c string) (result MyPopin) {
	if err := mongodb.FindA(database, collection, bson.M{"campaignID": c}, &result); err != nil {
		fmt.Println(c, "查询失败",err)
	}
	return
}

//更新popin的消耗
func (mgoPopin *MgoPopin) UpdatePopin(c string) (err error) {
	now := time.Now().Format("2006-01-02 15:04:05")
	money := mgoPopin.GetMoney(mgoPopin.Cookie, mgoPopin.Account, c)
	var result MyPopin
	if err = mongodb.FindA(database, collection, bson.M{"campaignID": c}, &result); err != nil {
		fmt.Println(c, "不存在",err)
		return
	}
	if err = mongodb.Update(database, collection, bson.M{"campaignID": c}, bson.M{"$set": bson.M{"totalmoney": money, "date": now}}); err != nil {
		fmt.Println(c, "更新失败",err)
		return
	}

	f := reflect.ValueOf(&result.TotalMoney)
	f.Elem().SetInt(int64(money))
	return mongodb.Insert(database, c, &result)
}

//查询popin群活动的总消耗，并发获取，协程计算
func (mgoPopin *MgoPopin) FindAllMoney(c []string) (sumMoney int){
	ttmoney = make(chan int)
	defer close(ttmoney)
	go func(){
		for {
			select{
			case n := <-ttmoney:
				sumMoney += n
			default:

			}
		}
	}()

	wg.Add(len(c))
	for _ , v := range c {
		go func(){
			defer wg.Done()
			ttmoney <- mgoPopin.FindPoin(v).TotalMoney
		}()
	}
	wg.Wait()

	return
}

//获取最新POP的消耗通过数据库
func (mgoPopin *MgoPopin) GetNewestMoney(c []string) (money int){
	for _,ca := range c {
		var result []MyPopin
		err := mongodb.FindNewest(database, ca,&result)
		if err != nil {
			fmt.Println("获取POP最新消耗失败")
		}else{
			money += result[0].TotalMoney
		}
	}
	return
}

//获取Mark时间节点的消耗
func (mgoPopin *MgoPopin) GetMarkMoney(a []string) (money int) {
	marktime := bson.MongoTimestamp(function.GetMarkTime())
	for _,v :=range a {
		var result []MyPopin
		err := mongodb.FindMark(database, v,marktime,&result)
		if err != nil {
			fmt.Printf("获取%s消耗失败\n",time.Unix(int64(marktime),0))
		}else{
			money += result[0].TotalMoney
		}
	}
	return
}