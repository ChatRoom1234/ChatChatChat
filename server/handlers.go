package server

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/ChatRoom1234/ChatChatChat/db"
)

type authPair struct {
	Login    string `json:"login,omitempty"`
	Password string `json:"password,omitempty"`
}

type responseJSON struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type responseArrayJSON struct {
	Type     string      `json:"type"`
	Messages [100]string `json:"messages,omitempty"`
	Message  string      `json:"message",omitempty`
}

func handlerDefault(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	http.Error(w, "Not found", 404)
	return
}

func handlerAuth(w http.ResponseWriter, req *http.Request) {
	// accepts json
	buf := new(bytes.Buffer)
	buf.ReadFrom(req.Body)

	var dat authPair

	err := json.Unmarshal(buf.Bytes(), &dat)
	if err != nil || dat.Login == "" || dat.Password == "" {
		log.Printf("error parsing json: %v ", err)
		http.Error(w, "Bad request", 400)
		return
	}

	userID := db.ValidateUser(dat.Login, dat.Password)

	var response *responseJSON
	if userID > 0 {
		accessKey := db.CreateKey(userID)
		response = &responseJSON{Type: "access_key", Message: accessKey}
	} else {
		log.Printf("auth error")
		response = &responseJSON{Type: "error", Message: "wrong login/password"}
	}

	resp, _ := json.Marshal(response)
	w.Write(resp)
}

func handlerReg(w http.ResponseWriter, req *http.Request) {
	// accepts json
	buf := new(bytes.Buffer)
	buf.ReadFrom(req.Body)

	// var dat map[string]interface{}
	var dat authPair

	err := json.Unmarshal(buf.Bytes(), &dat)
	if err != nil || dat.Login == "" || dat.Password == "" {
		log.Printf("error parsing json: %v ", err)
		http.Error(w, "Bad request", 400)
		return
	}

	userID, err := db.CreateUser(dat.Login, dat.Password)

	var response *responseJSON
	if err != nil {
		response = &responseJSON{Type: "error", Message: err.Error()}
	} else {
		accessKey := db.CreateKey(userID)
		response = &responseJSON{Type: "access_key", Message: accessKey}
	}

	resp, _ := json.Marshal(response)
	w.Write(resp)
}

func handlerHistory(w http.ResponseWriter, req *http.Request) {
	messages, err := db.GetHistory()

	var response *responseArrayJSON
	if err != nil {
		response = &responseArrayJSON{Type: "error", Message: err.Error()}
	} else {
		response = &responseArrayJSON{Type: "messages", Messages: messages}
	}

	resp, _ := json.Marshal(response)
	w.Write(resp)
}
