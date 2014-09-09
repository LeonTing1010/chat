package handlers

import (
	"encoding/json"
	. "github.com/paulbellamy/mango"
	"github.com/sunfmin/mangotemplate"
	// "io/ioutil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"net/http"
)

const (
	URL = "localhost:27017"
)

type RenderData struct {
	Id            string
	WebSocketHost string
}

//Result message
type Result struct {
	//{“code”:”0”,”info”:”ok”}
	Code int
	Info string
	Host string
	Id   string
}

func Home(env Env) (status Status, headers Headers, body Body) {
	mangotemplate.ForRender(env, "chats/home", nil)
	headers = Headers{}
	return
}

func Register(env Env) (status Status, headers Headers, body Body) {
	env.Logger().Println("Got a", env.Request().Method, "request for", env.Request().RequestURI)
	r := env.Request()
	headers = Headers{}
	decoder := json.NewDecoder(r.Body)

	var user User
	err := decoder.Decode(&user)
	if err != nil {
		env.Logger().Println("Got a err", r.Method, "read request body", err.Error())
		result, _ := json.Marshal(Result{Code: 1001, Info: err.Error(), Host: r.Host})
		body = Body(result)
		env.Logger().Println("Send a err", r.Method, "response body", body)
		return
	}
	user.Save()
	result, _ := json.Marshal(Result{Code: 0, Info: "ok", Host: r.Host, Id: user.Email})
	body = Body(result)
	env.Logger().Println("Send a ", r.Method, "response body", body)
	return
}

func Join(env Env) (status Status, headers Headers, body Body) {
	env.Logger().Println("Got a", env.Request().Method, "request for", env.Request().RequestURI)
	// email := env.Request().FormValue("email")
	// if email == "" {
	// 	return Redirect(http.StatusFound, "/")
	// }

	r := env.Request()
	// mangotemplate.ForRender(env, "chats/room", &RenderData{Id: email, WebSocketHost: r.Host})
	headers = Headers{}
	decoder := json.NewDecoder(r.Body)

	var user User
	err := decoder.Decode(&user)
	if err != nil {
		env.Logger().Println("Got a err", r.Method, "read request body", err.Error())
		result, _ := json.Marshal(Result{Code: 1001, Info: err.Error(), Host: r.Host})
		body = Body(result)
		env.Logger().Println("Send a err", r.Method, "response body", body)
		return
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
	collection.RemoveAll(bson.M{"email": nil})
	dbUser := User{}
	collection.Find(bson.M{"email": user.Email}).One(&dbUser)
	env.Logger().Println("Find in mydb", "user", "collection", dbUser.Nickname)
	if dbUser.Email == user.Email && dbUser.Password == user.Password {
		result, _ := json.Marshal(Result{Code: 0, Info: "ok", Host: r.Host, Id: dbUser.Email})
		body = Body(result)
		env.Logger().Println("Send a ", r.Method, "response body", body)
		// 存入cookie,使用cookie存储
		cookie := http.Cookie{Name: "userId", Value: dbUser.Email, Path: "/"}
		env.Request().AddCookie(&cookie)
		return
	}
	result, _ := json.Marshal(Result{Code: 1002, Info: "user not register or password wrong", Host: r.Host})
	body = Body(result)
	env.Logger().Println("Send a err", r.Method, "response body", body)

	return
}
