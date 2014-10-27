package main

import (
	"chat/handlers"
	"chat/wscon"
	"code.google.com/p/go.net/websocket"
	. "github.com/paulbellamy/mango"
	"net/http"
)

func main() {
	// l, r := handlers.LayoutAndRenderer()
	s := new(Stack)
	// s.Middleware(l, r)
	http.Handle("/chat", websocket.Handler(wscon.BuildConnection))

	http.HandleFunc("/register", s.HandlerFunc(handlers.Register))
	http.HandleFunc("/login", s.HandlerFunc(handlers.Join))
	http.HandleFunc("/update", s.HandlerFunc(handlers.UpdateUser))
	http.HandleFunc("/findfriend", s.HandlerFunc(handlers.FindFriend))
	http.HandleFunc("/addfriend", s.HandlerFunc(handlers.AddFriend))
	http.HandleFunc("/updatefriend", s.HandlerFunc(handlers.UpdateFriend))
	http.HandleFunc("/history", s.HandlerFunc(handlers.HistoryMsg))
	http.HandleFunc("/", s.HandlerFunc(handlers.Home))
	http.HandleFunc("/public/", assetsHandler)

	go wscon.InitChatRoom()

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}

func assetsHandler(w http.ResponseWriter, r *http.Request) {
	assetPath := "src/chatour/" + r.URL.Path[len("/"):]
	//log.Println(assetPath)
	http.ServeFile(w, r, assetPath)
}
