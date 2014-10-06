package main

import (
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/jelinden/rssFetcher/rss"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"net/http"
	"strings"
	"time"
)

var mongoAddress = flag.String("address", "localhost", "mongo address")
var session *mgo.Session

func wsHandler(ws *websocket.Conn) {
	defer func() {
		fmt.Printf(time.Now().Format("2006-01-02 15:04:05 -0700") + " Connection closed\n")
		ws.Close()
	}()
	msg := make([]byte, 1024)
	for {
		n, err := ws.Read(msg)
		if err != nil {
			fmt.Printf("Connection closed %s\n", err)
			break
		}
		fmt.Printf(time.Now().Format("2006-01-02 15:04:05 -0700")+" Receive: %s, %s\n", msg[:n], ws.Request().RemoteAddr)
		if strings.HasPrefix(string(msg[:n]), "c/") {
			saveClick(strings.Replace(string(msg[:n]), "c/", "", -1))
		} else if strings.Contains(ws.Request().RequestURI, "/fi/") {
			fetchRssItems(ws, 1)
		} else if strings.Contains(ws.Request().RequestURI, "/en/") {
			fetchRssItems(ws, 2)
		} else if strings.Contains(ws.Request().RequestURI, "/sv/") {
			fetchRssItems(ws, 3)
		} else {
			fetchRssItems(ws, 1)
		}
	}
	fmt.Printf(time.Now().Format("2006-01-02 15:04:05 -0700") + " => closing connection\n")
	ws.Close()
}

func main() {
	flag.Parse()
	fmt.Println("mongoAddress " + *mongoAddress)
	var err error
	if session, err = mgo.Dial(*mongoAddress); err != nil {
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	http.Handle("/websocket/", websocket.Handler(wsHandler))
	http.Handle("/websocket/fi/", websocket.Handler(wsHandler))
	http.Handle("/websocket/en/", websocket.Handler(wsHandler))
	http.Handle("/websocket/sv/", websocket.Handler(wsHandler))
	http.Handle("/", http.FileServer(http.Dir(".")))
	http.Handle("/fi/", http.StripPrefix("/fi/", http.FileServer(http.Dir("."))))
	http.Handle("/sv/", http.StripPrefix("/en/", http.FileServer(http.Dir("."))))
	http.Handle("/en/", http.StripPrefix("/sv/", http.FileServer(http.Dir("."))))
	http.ListenAndServe(":9100", nil)
}

func getFeedTitles(session *mgo.Session, language int) []rss.Item {
	result := []rss.Item{}
	c := session.DB("uutispuro").C("rss")
	err := c.Find(bson.M{"language": language}).Sort("-date").Limit(40).All(&result)
	if err != nil {
		fmt.Println("Fatal error " + err.Error())
	}
	return result
}

func saveClick(id string) {
	c := session.DB("uutispuro").C("rss")
	type M map[string]interface{}
	_, err := c.UpsertId(bson.ObjectIdHex(id), M{"$inc": M{"clicks": 1}})
	if err != nil {
		fmt.Println(err.Error())
	}
}

func fetchRssItems(ws *websocket.Conn, lang int) {
	doc := map[string]interface{}{"d": getFeedTitles(session, lang)}
	if data, err := json.Marshal(doc); err != nil {
		log.Printf(time.Now().Format("2006-01-02 15:04:05 -0700")+" Error marshalling json: %v", err)
	} else {
		ws.Write(data)
	}
}
