package wscon

import (
	// "chatour/libs"
	"chatour/handlers"
	"chatour/util"
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"strings"
)

const (
	MAX_SEND = 256
)

var runningActiveRoom *ActiveRoom = &ActiveRoom{}

func BuildConnection(ws *websocket.Conn) {
	userID := ws.Request().URL.Query().Get("id")

	if userID == "" {
		return
	}
	url := handlers.URL
	session, err := mgo.Dial(url) //连接数据库
	if err != nil {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	db := session.DB("mydb")   //数据库名称
	collection := db.C("user") //如果该集合已经存在的话，则直接返回
	dbUser := handlers.User{}
	collection.Find(bson.M{"Email": userID}).One(&dbUser)
	onlineUser := &OnlineUser{
		InRoom:     runningActiveRoom,
		Connection: ws,
		Send:       make(chan handlers.Message, MAX_SEND),
		UserInfo:   &dbUser,
	}
	runningActiveRoom.OnlineUsers[userID] = onlineUser
	log.Println(userID + " login")
	go onlineUser.PushToClient()
	onlineUser.PullFromClient()

	onlineUser.killUserResource()
}

type ActiveRoom struct {
	OnlineUsers map[string]*OnlineUser
	Broadcast   chan handlers.Message
	CloseSign   chan bool
}

type OnlineUser struct {
	InRoom     *ActiveRoom
	Connection *websocket.Conn
	UserInfo   *handlers.User
	Send       chan handlers.Message
}

type TextMessage struct {
	Content string
	Time    string
	To      string //email
}

func InitChatRoom() {
	runningActiveRoom = &ActiveRoom{
		OnlineUsers: make(map[string]*OnlineUser),
		Broadcast:   make(chan handlers.Message),
		CloseSign:   make(chan bool),
	}
	go runningActiveRoom.run()
}

// Core function of room
func (this *ActiveRoom) run() {
	for {
		select {
		case b := <-this.Broadcast:
			if b.MType == handlers.O2M_MTYPE {
				if len(b.From) <= 0 || !strings.Contains(b.From, "@") {
					log.Println("message formate error: message.From is not nil:" + b.MType + " " + b.From)
					continue
				}
				yes, _ := util.Contain(this.OnlineUsers, b.From)
				if !yes {
					log.Println(b.From + " offline,save o2m message in history")
					b.Save()
					continue
				}
				curUser := this.OnlineUsers[b.From].UserInfo
				friends := curUser.FindFriends()
				for _, online := range this.OnlineUsers {
					yes, _ := util.Contain(online.UserInfo.Email, friends)
					if yes {
						online.Send <- b
					} else {
						log.Println("you have no friend:" + curUser.Email)
					}
				}

			}
			if b.MType == handlers.O2O_MTYPE {
				yes, _ := util.Contain(this.OnlineUsers, b.From)
				if !yes {
					log.Println(b.From + " offline,save o2o message in history")
					b.Save()
					continue
				}
				this.OnlineUsers[b.Msg.To].Send <- b
			}

		case c := <-this.CloseSign:
			if c == true {
				close(this.Broadcast)
				close(this.CloseSign)
				return
			}
		}
	}
}

func (this *OnlineUser) PullFromClient() {
	for {
		var content string
		err := websocket.Message.Receive(this.Connection, &content)
		// If user closes or refreshes the browser, a err will occur
		if err != nil {
			panic(err)
			return
		}
		log.Println("receive msg:" + content)
		var m handlers.Message
		if err := json.Unmarshal([]byte(content), &m); err != nil {
			log.Println("err decode message to json :" + err.Error())
		}
		m.Msg.Time = util.HumanCreatedAt()
		log.Println("message receive time:" + m.Msg.Time)
		this.InRoom.Broadcast <- m
	}
}

func (this *OnlineUser) PushToClient() {
	for b := range this.Send {
		err := websocket.JSON.Send(this.Connection, b)
		if err != nil {
			break
		}
	}
}

func (this *OnlineUser) killUserResource() {
	this.Connection.Close()
	delete(this.InRoom.OnlineUsers, this.UserInfo.Email)
	close(this.Send)
}

func (this *ActiveRoom) GetOnlineUsers() (users []*handlers.User) {
	for _, online := range this.OnlineUsers {
		users = append(users, online.UserInfo)
	}
	return
}
