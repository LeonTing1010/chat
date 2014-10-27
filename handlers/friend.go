package handlers

import (
	"chat/util"
	"encoding/json"
	. "github.com/paulbellamy/mango"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

const (
	O_FTYPE = "ordinary"
	C_FTYPE = "close"
)

type Friend struct {
	FType string //O_FTYPE or C_FTYPE
	From  string
	To    string
	Time  string
}

func UpdateFriend(env Env) (status Status, headers Headers, body Body) {
	env.Logger().Println("Got a", env.Request().Method, "request for", env.Request().RequestURI)
	r := env.Request()
	headers = Headers{}
	var friend Friend
	requestBody := make([]byte, r.ContentLength)
	_, err := r.Body.Read(requestBody)

	if err != nil {
		env.Logger().Println("Got a err", r.Method, "read request body", err.Error())
		result, _ := json.Marshal(Result{Code: 1001, Info: err.Error(), Host: r.Host})
		body = Body(result)
		env.Logger().Println("Send a err", r.Method, "response body", body)
		return
	} else {
		json.Unmarshal(requestBody, &friend)
		env.Logger().Println("Got a err", r.Method, "request body", string(requestBody))
	}
	url := URL
	session, err := mgo.Dial(url) //连接数据库
	if err != nil {
		env.Logger().Println("Got a err", r.Method, "mgo Dial ", url)
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	env.Logger().Println("Connect to ", "mgo", "url", url)
	db := session.DB("mydb")     //数据库名称
	collection := db.C("friend") //如果该集合已经存在的话，则直接返回
	dbFriend := Friend{}
	if err := collection.Update(bson.M{"from": friend.From, "to": friend.To}, bson.M{"$set": bson.M{"ftype": C_FTYPE, "time": util.HumanCreatedAt()}}); err != nil {
		result, _ := json.Marshal(Result{Code: 1000, Info: "not found friend", Host: r.Host})
		body = Body(result)
		env.Logger().Println("Send a err ", r.Method, "response body", body)
		return
	}
	result, _ := json.Marshal(Result{Code: 0, Info: "ok", Host: r.Host, Id: dbFriend.From})
	body = Body(result)
	env.Logger().Println("Send a ", r.Method, "response body", body)
	return

}

func AddFriend(env Env) (status Status, headers Headers, body Body) {
	env.Logger().Println("Got a", env.Request().Method, "request for", env.Request().RequestURI)
	r := env.Request()
	url := URL
	session, err := mgo.Dial(url) //连接数据库
	if err != nil {
		env.Logger().Println("Got a err", r.Method, "mgo Dial ", url)
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	env.Logger().Println("Connect to ", "mgo", "url", url)
	db := session.DB("mydb")     //数据库名称
	collection := db.C("friend") //如果该集合已经存在的话，则直接返回
	query := r.URL.Query()
	from := query.Get("id")
	to := query.Get("friend")
	ftype := query.Get("ftype")
	if ftype == "" {
		ftype = O_FTYPE
	}
	friend := Friend{FType: ftype, From: from, To: to, Time: util.HumanCreatedAt()}
	err = collection.Insert(&friend) //一次可以插入多个对象
	if err != nil {
		env.Logger().Println("Got a err", r.Method, "mgo insert friend ", db.Name)
		panic(err)
	}
	result, _ := json.Marshal(Result{Code: 0, Info: "ok", Host: r.Host, Id: from})
	body = Body(result)
	env.Logger().Println("Send a ", r.Method, "response body", body)
	return
}
func FindFriend(env Env) (status Status, headers Headers, body Body) {
	env.Logger().Println("Got a", env.Request().Method, "request for", env.Request().RequestURI)
	r := env.Request()
	url := URL
	session, err := mgo.Dial(url) //连接数据库
	if err != nil {
		env.Logger().Println("Got a err", r.Method, "mgo Dial ", url)
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	env.Logger().Println("Connect to ", "mgo", "url", url)
	db := session.DB("mydb")   //数据库名称
	collection := db.C("user") //如果该集合已经存在的话，则直接返回
	friends := []User{}
	userID := r.URL.Query().Get("id")
	iter := collection.Find(bson.M{"email": bson.M{"$ne": userID}}).Limit(10).Iter()
	if err := iter.All(&friends); err != nil {
		result, _ := json.Marshal([]User{})
		body = Body(result)
		env.Logger().Println("Send a err ", r.Method, "response body", body)
		return
	}
	result, _ := json.Marshal(friends)
	body = Body(result)
	env.Logger().Println("Send a  ", r.Method, "response body", body)
	return
}
