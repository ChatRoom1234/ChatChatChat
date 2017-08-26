package server

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/ChatRoom1234/ChatChatChat/chatroom"
)

var addr = flag.String("addr", ":8080", "http service address")

// flag.Parse()

func urlSolver(room *chatroom.Room, w http.ResponseWriter, req *http.Request) {
	fmt.Println("Requested url is: ", req.URL.Path)
	switch req.URL.Path {
	case "/":
		http.ServeFile(w, req, "server/home.html")
	case "/reg":
		handlerReg(w, req)
	case "/auth":
		handlerAuth(w, req)
	case "/history":
		handlerHistory(w, req)
	case "/ws":
		chatroom.ServeWs(room, w, req)
	default:
		handlerDefault(w, req)
	}
}

// NewServer creates a chatroom server
func NewServer(room *chatroom.Room) error {
	return http.ListenAndServe(*addr, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		urlSolver(room, w, r)
	}))
}
