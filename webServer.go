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
)

var mongoAddress = flag.String("address", "localhost", "mongo address")
var session *mgo.Session

func wsHandler(ws *websocket.Conn) {

	defer func() {
		fmt.Printf("Connection closed")
		ws.Close()
	}()
	msg := make([]byte, 1024)
	for {
		n, err := ws.Read(msg)
		if err != nil {
			fmt.Printf("Connection closed %s", err)
			break
		}
		fmt.Printf("Receive: %s, %s\n", msg[:n], ws.Request().RemoteAddr)

		doc := map[string]interface{}{"d": getFeedTitles(session)}

		if data, err := json.Marshal(doc); err != nil {
			log.Printf("Error marshalling json: %v", err)
		} else {
			ws.Write(data)
			if err != nil {
				fmt.Printf("Connection closed %s", err)
				break
			}
		}
	}
	fmt.Printf(" => closing connection\n")
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
	http.Handle("/websocket", websocket.Handler(wsHandler))
	http.Handle("/", http.FileServer(http.Dir(".")))
	http.ListenAndServe(":9100", nil)
}

func getFeedTitles(session *mgo.Session) []rss.Item {
	result := []rss.Item{}
	c := session.DB("uutispuro").C("rss")
	err := c.Find(bson.M{}).Sort("-date").Limit(40).All(&result)
	if err != nil {
		fmt.Println("Fatal error " + err.Error())
	}
	return result
}
