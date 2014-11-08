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
	"text/template"
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
			fetchRssItems(ws, 2)
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
		fmt.Println("closing up, no connection to mongo")
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	fmt.Println("adding routes")
	http.Handle("/websocket/", websocket.Handler(wsHandler))
	http.HandleFunc("/", rootHandler)
	fmt.Println("starting web server")
	http.ListenAndServe(":9100", nil)
}

func getFeedTitles(session *mgo.Session, language int, limit int) []rss.Item {
	result := []rss.Item{}
	s := session.Clone()
	c := s.DB("uutispuro").C("rss")
	err := c.Find(bson.M{"language": language}).Sort("-date").Limit(limit).All(&result)
	if err != nil {
		fmt.Println("Fatal error " + err.Error())
	}
	return addCategoryShowNames(result)
}

func saveClick(id string) {
	s := session.Clone()
	c := s.DB("uutispuro").C("rss")
	type M map[string]interface{}
	_, err := c.UpsertId(bson.ObjectIdHex(id), M{"$inc": M{"clicks": 1}})
	if err != nil {
		fmt.Println(err.Error())
	}
}

func fetchRssItems(ws *websocket.Conn, lang int) {
	doc := map[string]interface{}{"d": getFeedTitles(session, lang, 40)}
	if data, err := json.Marshal(doc); err != nil {
		log.Printf("Error marshalling json: %v", err)
	} else {
		ws.Write(data)
	}
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("uri " + r.RequestURI)
	var content []byte = nil
	if strings.HasPrefix(r.RequestURI, "/fi") {
		htmlTemplateFi(w, r)
	} else if strings.EqualFold(r.RequestURI, "/") || strings.HasPrefix(r.RequestURI, "/en") {
		htmlTemplateEn(w, r)
	} else if strings.HasSuffix(r.RequestURI, "/uutispuro.css") {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d, public, must-revalidate, proxy-revalidate", 60*60*24*7*4))
		content = openFileGzipped("uutispuro.css")
	} else if strings.HasSuffix(r.RequestURI, "/uutispuro.js") {
		w.Header().Set("Content-Type", "application/javascript")
		w.Header().Set("Content-Encoding", "gzip")
		w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d, public, must-revalidate, proxy-revalidate", 60*60*24*7*4))
		content = openFileGzipped("uutispuro.js")
	} else if strings.HasSuffix(r.RequestURI, "/favicon.ico") {
		w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d, public, must-revalidate, proxy-revalidate", 60*60*24*7*4))
		content = openFile("img/favicon.ico")
	} else if strings.HasSuffix(r.RequestURI, ".png") {
		w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d, public, must-revalidate, proxy-revalidate", 60*60*24*7*4))
		w.Header().Set("Content-Type", "image/png")
		content = openFile("img" + r.RequestURI[strings.LastIndex(r.RequestURI, "/"):len(r.RequestURI)])
	} else if strings.HasSuffix(r.RequestURI, ".gif") {
		w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d, public, must-revalidate, proxy-revalidate", 60*60*24*7*4))
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

func htmlTemplateEn(w http.ResponseWriter, r *http.Request) {
	htmlTemplate(w, r, getFeedTitles(session, 2, 10))
}

func htmlTemplateFi(w http.ResponseWriter, r *http.Request) {
	htmlTemplate(w, r, getFeedTitles(session, 1, 10))
}

func htmlTemplate(w http.ResponseWriter, r *http.Request, items []rss.Item) {

	t, err := template.ParseFiles("index.html")
	if err != nil {
		log.Printf("Template gave: %s", err)
	}

	cangzip := strings.Index(r.Header.Get("Accept-Encoding"), "gzip") > -1
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if cangzip {
		gw := gzip.NewWriter(w)
		w.Header().Set("Vary", "Accept-Encoding")
		w.Header().Set("Content-Encoding", "gzip")
		t.Execute(gw, items)
		gw.Close()
	} else {
		t.Execute(w, items)
	}
}

func addCategoryShowNames(items []rss.Item) []rss.Item {
	for i := range items {
		if items[i].Category.Name == "Asuminen" {
			items[i].Category.StyleName = "Koti"
		} else if items[i].Category.Name == "IT ja media" {
			items[i].Category.StyleName = "Digi"
		} else if items[i].Category.Name == "Naiset ja muoti" {
			items[i].Category.StyleName = "Naisetjamuoti"
		} else if items[i].Category.Name == "TV ja elokuvat" {
			items[i].Category.StyleName = "Elokuvat"
		} else {
			items[i].Category.StyleName = items[i].Category.Name
		}

	}
	return items
}
