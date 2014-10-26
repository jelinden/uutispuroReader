package main

import (
	"bytes"
	"code.google.com/p/go.net/websocket"
	"compress/gzip"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/jelinden/rssFetcher/rss"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

var mongoAddress = flag.String("address", "localhost", "mongo address")
var session *mgo.Session

func wsHandler(ws *websocket.Conn) {
	defer func() {
		log.Printf(" Connection closed\n")
		ws.Close()
	}()
	msg := make([]byte, 1024)
	for {
		n, err := ws.Read(msg)
		if err != nil {
			log.Printf("Connection closed %s\n", err)
			break
		}
		log.Println(ws.Request().RequestURI)
		if strings.HasPrefix(string(msg[:n]), "c/") {
			saveClick(strings.Replace(string(msg[:n]), "c/", "", -1))
		} else if strings.HasSuffix(ws.Request().RequestURI, "/fi/") {
			fetchRssItems(ws, 1)
		} else if strings.HasSuffix(ws.Request().RequestURI, "/en/") {
			fetchRssItems(ws, 2)
		} else if strings.HasSuffix(ws.Request().RequestURI, "/sv/") {
			fetchRssItems(ws, 3)
		} else if strings.EqualFold(ws.Request().RequestURI, "/websocket/") {
			fetchRssItems(ws, 1)
		}
	}
	log.Printf(" => closing connection\n")
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
	http.HandleFunc("/", rootHandler)
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
		log.Printf("Error marshalling json: %v", err)
	} else {
		ws.Write(data)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("uri " + r.RequestURI)
	var content []byte = nil
	if strings.EqualFold(r.RequestURI, "/") || strings.EqualFold(r.RequestURI, "/fi/") || strings.EqualFold(r.RequestURI, "/en/") || strings.EqualFold(r.RequestURI, "/sv/") {
		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("Content-Encoding", "gzip")
		content = openFileGzipped("index.html")
	} else if strings.EqualFold(r.RequestURI, "/uutispuro.css") {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		w.Header().Set("Content-Encoding", "gzip")
		content = openFileGzipped("uutispuro.css")
	} else if strings.EqualFold(r.RequestURI, "/uutispuro.js") {
		w.Header().Set("Content-Type", "application/javascript")
		w.Header().Set("Content-Encoding", "gzip")
		content = openFileGzipped("uutispuro.js")
	} else if strings.HasSuffix(r.RequestURI, "/favicon.ico") {
		content = openFile("img/favicon.ico")
	} else if strings.HasSuffix(r.RequestURI, ".png") {
		w.Header().Set("Content-Type", "image/png")
		content = openFile("img" + r.RequestURI[strings.LastIndex(r.RequestURI, "/"):len(r.RequestURI)])
	} else if strings.HasSuffix(r.RequestURI, ".gif") {
		w.Header().Set("Content-Type", "image/gif")
		content = openFile("img" + r.RequestURI[strings.LastIndex(r.RequestURI, "/"):len(r.RequestURI)])
	}
	fmt.Fprintf(w, "%s", content)
}

func openFile(fileName string) []byte {
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Println("Could not open file.", err)
	}
	return content
}

func openFileGzipped(fileName string) []byte {
	content, err := ioutil.ReadFile(fileName)
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	w.Write(content)
	w.Close()
	if err != nil {
		log.Println("Could not open file.", err)
	}
	return b.Bytes()
}
