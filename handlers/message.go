package handlers

import (
	// "chatour/libs"
	"encoding/json"
	. "github.com/paulbellamy/mango"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
)

const (
	O2M_MTYPE = "o2m"
	O2O_MTYPE = "o2o"
)

//{"mtype":"o2m","msg":{"content":"hello","time":"2014-01-02 15:04:05","to":"userid"},"from":"userid"}
type Message struct {
	MType string // o2o or o2m
	Msg   TextMessage
	From  string //email
}

type TextMessage struct {
	Content string
	Time    string
	To      string //email
}

func (this *Message) Save() {
	session, err := mgo.Dial(URL) //连接数据库
	if err != nil {
		log.Println("Got a err", "Save", "mgo Dial ", URL)
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	log.Println("Connect to ", "mgo", "url", URL)
	db := session.DB("mydb")      //数据库名称
	collection := db.C("message") //如果该集合已经存在的话，则直接返回
	log.Println("Insert a new message", this.MType, "mgo ", URL)
	err = collection.Insert(&this) //一次可以插入多个对象
	if err != nil {
		panic(err)
	}
}

//historymsg?id=email
func HistoryMsg(env Env) (status Status, headers Headers, body Body) {
	env.Logger().Println("Got a", env.Request().Method, "request for", env.Request().RequestURI)
	r := env.Request()
	headers = Headers{}
	userId := r.URL.Query().Get("id")
	session, err := mgo.Dial(URL) //连接数据库
	if err != nil {
		env.Logger().Println("Got a err", r.Method, "mgo Dial ", URL)
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	env.Logger().Println("Connect to ", "mgo", "url", URL)
	db := session.DB("mydb")      //数据库名称
	collection := db.C("message") //如果该集合已经存在的话，则直接返回
	msgs := []Message{}
	iter := collection.Find(bson.M{"from": userId}).Iter()
	if err := iter.All(&msgs); err != nil {
		panic(err)
		return
	}
	result, _ := json.Marshal(msgs)
	body = Body(result)
	env.Logger().Println("Send a ", r.Method, "response body", body)
	return

}
