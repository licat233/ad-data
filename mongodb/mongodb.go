package mongodb

import (
	"ad-data/config"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
)

var globalS *mgo.Session

const (
	dbhost    = config.DBhost
	authdb    = config.DBname
	authuser  = config.DBUser
	authpass  = config.DBPwd
	timeout   = config.DBtimeout
	poollimit = config.DBpoollimit
)

func init() {
	dialInfo := &mgo.DialInfo{
		Addrs:     []string{dbhost},
		Timeout:   timeout,
		Source:    authdb,
		Username:  authuser,
		Password:  authpass,
		PoolLimit: poollimit,
	}
	s, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		log.Fatalf("MongoDB数据库连接失败: %s\n", err)
	}
	globalS = s
}

//db:操作的数据库
//collection:操作的文档(表)
//page:当前页面
//limit:每页的数量值
//query:查询条件
//selector:需要过滤的数据(projection)
//result:查询到的结果

//链接数据库
func Connect(db, collection string) (*mgo.Session, *mgo.Collection) {
	ms := globalS.Copy()
	c := ms.DB(db).C(collection)
	ms.SetMode(mgo.Monotonic, true)
	return ms, c
}

//插入
func Insert(db, collection string, doc interface{}) error {
	ms, c := Connect(db, collection)
	defer ms.Close()
	return c.Insert(doc)
}

//寻找
func FindOne(db, collection string, query, selector, result interface{}) error {
	ms, c := Connect(db, collection)
	defer ms.Close()
	return c.Find(query).Select(selector).One(result)
}
//test 查询title="标题",并且返回结果中去除`_id`字段
//var result Data
//err = db.FindOne(database, collection, bson.M{"title": "标题"}, bson.M{"_id":0}, &result)

func FindA(db, collection string, query, result interface{}) error {
	ms, c := Connect(db, collection)
	defer ms.Close()
	return c.Find(query).One(result)
}

//查询最新的数据 result必须为切片
func FindNewest(db, collection string,result interface{}) error {
	ms, c := Connect(db, collection)
	defer ms.Close()
	return c.Find(nil).Sort("-date").Limit(1).All(result)
}

//查询Marke时间的数据,一定要传入切片
func FindMark(db, collection string,marktime bson.MongoTimestamp,res interface{}) error{
	ms, c := Connect(db, collection)
	defer ms.Close()
	return c.Find(bson.M{"date":bson.M{"$gte":marktime}}).Limit(1).Sort("date").All(res)
}

//查询单条数据的ID
func FindID(db,collection string,find interface{}) (id string, err error){
	ms, c := Connect(db, collection)
	defer ms.Close()
	type ObjectID struct {
		ID bson.ObjectId `bson:"_id"`
	}
	obi := ObjectID{}
	err = c.Find(find).One(&obi)
	id = bson.ObjectId.Hex(obi.ID)
	return
}

func FindAll(db, collection string, query, selector, result interface{}) error {
	ms, c := Connect(db, collection)
	defer ms.Close()
	return c.Find(query).Select(selector).All(result)
}

//查询集合里所有的数据
func FindAllData(db, collection string,result interface{}) error {
	ms, c := Connect(db, collection)
	defer ms.Close()
	return c.Find(nil).Sort("-date").All(result)
}


//更新
func Update(db, collection string, selector, update interface{}) error {
	ms, c := Connect(db, collection)
	defer ms.Close()

	return c.Update(selector, update)
}

//更新，如果不存在就插入一个新的数据 `Upsert:true`
func Upsert(db, collection string, selector, update interface{}) error {
	ms, c := Connect(db, collection)
	defer ms.Close()

	_, err := c.Upsert(selector, update)
	return err
}

// `multi:true`
func UpdateAll(db, collection string, selector, update interface{}) error {
	ms, c := Connect(db, collection)
	defer ms.Close()

	_, err := c.UpdateAll(selector, update)
	return err
}

//删除
func Remove(db, collection string, selector interface{}) error {
	ms, c := Connect(db, collection)
	defer ms.Close()

	return c.Remove(selector)
}

func RemoveAll(db, collection string, selector interface{}) error {
	ms, c := Connect(db, collection)
	defer ms.Close()

	_, err := c.RemoveAll(selector)
	return err
}

//其它操作
//判断集合是否为空
func IsEmpty(db, collection string) bool {
	ms, c := Connect(db, collection)
	defer ms.Close()
	count, err := c.Count()
	if err != nil {
		log.Fatal(err)
	}
	return count == 0
}

//查找集合里面的文档数
func Count(db, collection string, query interface{}) (int, error) {
	ms, c := Connect(db, collection)
	defer ms.Close()
	return c.Find(query).Count()
}

//分页查询
func FindPage(db, collection string, page, limit int, query, selector, result interface{}) error {
	ms, c := Connect(db, collection)
	defer ms.Close()

	return c.Find(query).Select(selector).Skip(page * limit).Limit(limit).All(result)
}

