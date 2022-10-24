package main

type Message struct {
	Author  string
	Title   string
	Content string
}

type JSONGroup struct {
	Id   string
	Data any
}

type JSONSession struct {
	Id      string
	Data    SessionInfo
	Hook    string
	GroupID string
	Group   JSONGroup
}

type JSONTBCPush struct {
	Message Message
	Session JSONSession
}
