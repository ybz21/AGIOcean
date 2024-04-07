package models

import (
	"github.com/gorilla/websocket"
)

type Ability struct {
	Type  string `json:"type"`  // text-generation
	Model string `json:"model"` // qwen-7b
	//Price string `json:"price"` //计费方式
}

type Worker struct {
	ID         string    `json:"id"`
	Token      string    `json:"token"`
	IP         string    `json:"ip"`
	Online     bool      `json:"online"`
	Abilities  []Ability `json:"abilities"`
	ModelName  string    `json:"modelName"`
	Connection *websocket.Conn
}
