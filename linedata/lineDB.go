package linedata

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
	database   = "Line"
	collection = "account"
)

//linedata重点在于更新粉丝数fans
//MgoLine Mgo业务结构，备用的，不用纠结这是干嘛的
type MgoLine struct {
}

//Line 业务号结构，主要用于定义文档结构
type MyLine struct {
	//Id      bson.ObjectId `bson:"_id"`
	Name    string              `bson:"name"`
	Account string              `bson:"account"`
	Fans    int                 `bson:"fans"`
	Date    bson.MongoTimestamp `bson:"date"`
}

var (
	dbMutex sync.RWMutex
	ttfans  chan int
)

//操作时注意读写锁的使用，防止死锁、活锁、饥饿锁的发生
//主要逻辑：
//注册：存在则更新
//更新：更新之后在指定集合插入新的文档，集合不存在则创建
//注册
func (mgoLine *MgoLine) RegisterLine(allLine map[string]string) {
	wg.Add(len(allLine))
	for n, a := range allLine {
		go mgoLine.mgoInsert(n, a)
	}
	wg.Wait()
}

func (mgoLine *MgoLine) mgoInsert(n, a string) {
	//now := time.Now().Format("2006-01-02 15:04:05")
	defer wg.Done()
	fansCount := GetFans(a)
	l := MyLine{
		Name:    n,
		Account: a,
		Fans:    fansCount,
		//Date:    now,
	}
	//数据库加锁
	//dbMutex.Lock()
	//更新，如果不存在就插入一个新的数据
	if err := mongodb.Upsert(database, collection, bson.M{"account": a}, &l); err != nil {
		fmt.Println(a, "Line项目更新注册失败", err)
	}
	//////////////旧方案
	//if err := mongodb.FindA(database, collection, bson.M{"account": a}, &l);err == nil {
	//	//说明已经存在，更新即可
	//	err = mongodb.Update(database, collection, bson.M{"account": a}, bson.M{"$set": bson.M{"fans": fansCount, "date": now}})
	//	if err != nil {
	//		fmt.Println(a, "更新失败",err)
	//	}
	//	return
	//}
	//
	//if err := mongodb.Insert(database, collection, &l); err != nil {
	//	fmt.Println(a,"注册失败",err)
	//}
	//////////////////

	//注册单个line account文档对应的集合，不存在会自动创建
	//if err := mongodb.FindA(database, a, bson.M{"account": a}, &l);err == nil {
	//	//说明已经存在，更新即可
	//}
	if err := mongodb.Insert(database, a, &l); err != nil {
		fmt.Println(a, "LINE子项目注册失败", err)
	}
	//数据库解锁
	//dbMutex.Unlock()
	return
}

//FindLine 查询line的信息
func (mgoLine *MgoLine) FindLine(a string) (result MyLine) {
	if err := mongodb.FindA(database, collection, bson.M{"account": a}, &result); err != nil {
		fmt.Println(a, "查询失败", err)
	}
	return
}

//更新line的粉丝数
func (mgoLine *MgoLine) UpdateLine(a string) (err error) {
	now := time.Now().Format("2006-01-02 15:04:05")
	fansCount := GetFans(a)
	var result MyLine
	if err = mongodb.FindA(database, collection, bson.M{"account": a}, &result); err != nil {
		fmt.Println(a, "不存在", err)
		return
	}
	if err = mongodb.Update(database, collection, bson.M{"account": a}, bson.M{"$set": bson.M{"fans": fansCount, "date": now}}); err != nil {
		fmt.Println(a, "更新失败", err)
		return
	}

	f := reflect.ValueOf(&result.Fans)
	f.Elem().SetInt(int64(fansCount))
	return mongodb.Insert(database, a, &result)
}

//查询line群活动的总消耗，并发获取，协程计算
func (mgoLine *MgoLine) FindAllFans(a map[string]string) (sumFans int) {
	ttfans = make(chan int)
	defer close(ttfans)
	go func() {
		for {
			select {
			case n := <-ttfans:
				sumFans += n
			default:

			}
		}
	}()

	wg.Add(len(a))
	for _, v := range a {
		go func() {
			defer wg.Done()
			ttfans <- mgoLine.FindLine(v).Fans
		}()
	}
	wg.Wait()

	return
}

//获取最新Line的粉丝数通过数据库
func (mgoLine *MgoLine) GetNewestFans(a map[string]string) (fans int) {
	for _, v := range a {
		var result []MyLine
		err := mongodb.FindNewest(database, v, &result)
		if err != nil {
			fmt.Printf("获取LINE%s最新粉丝数失败\n",v)
		}else{
			fans += result[0].Fans
		}
	}
	return
}

//获取Mark时间节点的粉丝数
func (mgoLine *MgoLine) GetMarkFans(a map[string]string) (fans int) {
	marktime := bson.MongoTimestamp(function.GetMarkTime())
	for _, v := range a {
		var result []MyLine
		err := mongodb.FindMark(database, v, marktime, &result)
		if err != nil {
			fmt.Printf("获取%s粉丝数失败\n",time.Unix(int64(marktime),0))
		}else{
			fans += result[0].Fans
		}
	}
	return
}
