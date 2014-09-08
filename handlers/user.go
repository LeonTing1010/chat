package handlers

import (
	"encoding/json"
	. "github.com/paulbellamy/mango"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
)

type User struct {
	//昵称，账号（邮箱），密码，性别，推荐人（待定）
	//'{"nickname":"zhangsan","email":"zhangsan@sina.com","password":"1234","sex":"male","referee":"lisi","phone":"10086"}'
	Nickname string
	Email    string
	Password string
	Sex      string
	Referee  string
	Phone    string
	Avatar   string
}

func (this *User) FindFriends() (friendIds []string) {

	session, err := mgo.Dial(URL) //连接数据库
	if err != nil {
		log.Println("Got a err", "FindFriends", "mgo Dial ", URL)
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	log.Println("Connect to ", "mgo", "url", URL)
	db := session.DB("mydb")     //数据库名称
	collection := db.C("friend") //如果该集合已经存在的话，则直接返回
	friends := []Friend{}
	iter := collection.Find(bson.M{"from": this.Email}).Iter()
	if err := iter.All(&friends); err != nil {
		panic(err)
		return
	}
	result := make([]string, len(friends), len(friends))
	for _, f := range friends {
		result = append(result, f.To)
	}
	return result
}
func (this *User) Save() {
	session, err := mgo.Dial(URL) //连接数据库
	if err != nil {
		log.Println("Got a err", "Save", "mgo Dial ", URL)
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	log.Println("Connect to ", "mgo", "url", URL)
	db := session.DB("mydb")   //数据库名称
	collection := db.C("user") //如果该集合已经存在的话，则直接返回
	if n, err := collection.Find(bson.M{"email": this.Email}).Count(); err != nil || n == 0 {
		panic(err)
		return
	}
	log.Println("Insert a new user", this.Email, "mgo ", URL)
	err = collection.Insert(&this) //一次可以插入多个对象
	if err != nil {
		panic(err)
	}
}

func UpdateUser(env Env) (status Status, headers Headers, body Body) {
	env.Logger().Println("Got a", env.Request().Method, "request for", env.Request().RequestURI)
	r := env.Request()
	headers = Headers{}
	var user User
	requestBody := make([]byte, r.ContentLength)
	_, err := r.Body.Read(requestBody)

	if err != nil {
		env.Logger().Println("Got a err", r.Method, "read request body", err.Error())
		result, _ := json.Marshal(Result{Code: 1001, Info: err.Error(), Host: r.Host})
		body = Body(result)
		env.Logger().Println("Send a err", r.Method, "response body", body)
		return
	} else {
		json.Unmarshal(requestBody, &user)
		env.Logger().Println("Got a err", r.Method, "request body", string(requestBody))
	}
	session, err := mgo.Dial(URL) //连接数据库
	if err != nil {
		env.Logger().Println("Got a err", r.Method, "mgo Dial ", URL)
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	env.Logger().Println("Connect to ", "mgo", "url", URL)
	db := session.DB("mydb")   //数据库名称
	collection := db.C("user") //如果该集合已经存在的话，则直接返回
	if err := collection.Update(bson.M{"email": user.Email}, bson.M{"$set": bson.M{"nickname": user.Nickname, "password": user.Password, "avatar": user.Avatar}}); err != nil {
		result, _ := json.Marshal(Result{Code: 1000, Info: "not found user", Host: r.Host})
		body = Body(result)
		env.Logger().Println("Send a err ", r.Method, "response body", body)
		return
	}
	result, _ := json.Marshal(Result{Code: 0, Info: "ok", Host: r.Host, Id: user.Email})
	body = Body(result)
	env.Logger().Println("Send a ", r.Method, "response body", body)
	return

}
