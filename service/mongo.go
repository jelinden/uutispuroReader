package service

import (
	"fmt"
	mgo "gopkg.in/mgo.v2"
	"time"
)

type Sessions struct {
	Mongo *mgo.Session
}

func NewSessions() *Sessions {
	return &Sessions{}
}

func (s *Sessions) Init() {
	s.Mongo = s.createSession(Conf.MongoUrl)
}

func (s *Sessions) Close() {
	s.Mongo.Close()
}

func (s *Sessions) createSession(url string) *mgo.Session {
	maxWait := time.Duration(5 * time.Second)
	session, err := mgo.DialWithTimeout(url, maxWait)
	if err != nil {
		fmt.Println("connection lost")
	}
	session.SetMode(mgo.Monotonic, true)
	return session
}
