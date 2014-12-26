package main

import (
	"bytes"
	"code.google.com/p/go.net/websocket"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/jelinden/rssFetcher/rss"
	"github.com/jelinden/uutispuroReader/service"
	"gopkg.in/mgo.v2/bson"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"
)

type Application struct {
	Sessions *service.Sessions
}

type Result struct {
	Items       []rss.Item
	Description string
	Lang        int
}

func NewApplication() *Application {
	return &Application{}
}

func (a *Application) Init() {
	a.Sessions = service.NewSessions()
	a.Sessions.Init()
}

func (a *Application) Close() {
	a.Sessions.Close()
}

func main() {
	app := NewApplication()
	app.Init()
	defer app.Close()

	fmt.Println("adding routes")
	http.Handle("/websocket/", websocket.Handler(app.WsHandler))
	http.HandleFunc("/", app.rootHandler)
	fmt.Println("starting web server")
	http.ListenAndServe(":9100", nil)
}

func (a *Application) WsHandler(ws *websocket.Conn) {
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
		log.Println(ws.Request().RemoteAddr, ws.Request().RequestURI)
		if strings.HasPrefix(string(msg[:n]), "c/") {
			a.saveClick(strings.Replace(string(msg[:n]), "c/", "", -1))
		} else if strings.HasPrefix(string(msg[:n]), "l/") {
			a.saveLike(strings.Replace(string(msg[:n]), "l/", "", -1))
		} else if strings.HasPrefix(string(msg[:n]), "u/") {
			a.saveUnlike(strings.Replace(string(msg[:n]), "u/", "", -1))
		} else if strings.HasSuffix(ws.Request().RequestURI, "/fi/") {
			a.fetchRssItems(ws, 1)
		} else if strings.HasSuffix(ws.Request().RequestURI, "/en/") {
			a.fetchRssItems(ws, 2)
		} else if strings.HasSuffix(ws.Request().RequestURI, "/sv/") {
			a.fetchRssItems(ws, 3)
		} else if strings.EqualFold(ws.Request().RequestURI, "/websocket/") {
			a.fetchRssItems(ws, 2)
		}
	}
	log.Printf(" => closing connection\n")
	ws.Close()
}

func (a *Application) getFeedTitles(language int, limit int) Result {
	result := []rss.Item{}
	s := a.Sessions.Mongo.Clone()
	c := s.DB("uutispuro").C("rss")
	err := c.Find(bson.M{"language": language}).Sort("-date").Limit(limit).All(&result)
	if err != nil {
		fmt.Println("Fatal error " + err.Error())
	}
	return a.addCategoryShowNamesAndMetaData(result, language)
}

func (a *Application) saveClick(id string) {
	s := a.Sessions.Mongo.Clone()
	c := s.DB("uutispuro").C("rss")
	type M map[string]interface{}
	_, err := c.UpsertId(bson.ObjectIdHex(id), M{"$inc": M{"clicks": 1}})
	if err != nil {
		fmt.Println(err.Error())
	}
}

func (a *Application) saveLike(id string) {
	s := a.Sessions.Mongo.Clone()
	c := s.DB("uutispuro").C("rss")
	type M map[string]interface{}
	_, err := c.UpsertId(bson.ObjectIdHex(id), M{"$inc": M{"likes": 1}})
	if err != nil {
		fmt.Println(err.Error())
	}
}

func (a *Application) saveUnlike(id string) {
	s := a.Sessions.Mongo.Clone()
	c := s.DB("uutispuro").C("rss")
	type M map[string]interface{}
	_, err := c.UpsertId(bson.ObjectIdHex(id), M{"$inc": M{"unlikes": 1}})
	if err != nil {
		fmt.Println(err.Error())
	}
}

func (a *Application) fetchRssItems(ws *websocket.Conn, lang int) {
	doc := map[string]interface{}{"d": a.getFeedTitles(lang, 45)}
	if data, err := json.Marshal(doc); err != nil {
		log.Printf("Error marshalling json: %v", err)
	} else {
		ws.Write(data)
	}
}

func (a *Application) rootHandler(w http.ResponseWriter, r *http.Request) {
	//pageNumber := a.getPage(r)
	//log.Println(r.RemoteAddr, r.RequestURI, pageNumber)
	var content []byte = nil
	if strings.HasPrefix(r.RequestURI, "/fi") {
		a.htmlTemplateFi(w, r)
	} else if strings.EqualFold(r.RequestURI, "/") || strings.HasPrefix(r.RequestURI, "/en") {
		a.htmlTemplateEn(w, r)
	} else if strings.HasSuffix(r.RequestURI, ".css") {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		w.Header().Set("Content-Encoding", "gzip")
		a.setHttpCacheHeaders(w.Header())
		content = a.openFileGzipped("styles" + r.RequestURI[strings.LastIndex(r.RequestURI, "/"):len(r.RequestURI)])
	} else if strings.HasSuffix(r.RequestURI, "/uutispuro-14.js") {
		w.Header().Set("Content-Type", "application/javascript")
		w.Header().Set("Content-Encoding", "gzip")
		a.setHttpCacheHeaders(w.Header())
		content = a.openFileGzipped("uutispuro-14.js")
	} else if strings.HasSuffix(r.RequestURI, "/favicon.ico") {
		a.setHttpCacheHeaders(w.Header())
		content = a.openFile("img/favicon.ico")
	} else if strings.HasSuffix(r.RequestURI, ".png") {
		a.setHttpCacheHeaders(w.Header())
		w.Header().Set("Content-Type", "image/png")
		content = a.openFile("img" + r.RequestURI[strings.LastIndex(r.RequestURI, "/"):len(r.RequestURI)])
	} else if strings.HasSuffix(r.RequestURI, ".gif") {
		w.Header().Set("Content-Type", "image/gif")
		a.setHttpCacheHeaders(w.Header())
		content = a.openFile("img" + r.RequestURI[strings.LastIndex(r.RequestURI, "/"):len(r.RequestURI)])
	}
	fmt.Fprintf(w, "%s", content)
}

func (a *Application) setHttpCacheHeaders(header http.Header) {
	header.Set("Cache-Control", fmt.Sprintf("max-age=%d, public, must-revalidate, proxy-revalidate", 60*60*24*7*4))
	header.Set("Last-Modified", time.Now().Format(http.TimeFormat))
	header.Set("Expires", time.Now().AddDate(60, 0, 0).Format(http.TimeFormat))
}

func (a *Application) openFile(fileName string) []byte {
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		log.Println("Could not open file.", err)
	}
	return content
}

func (a *Application) openFileGzipped(fileName string) []byte {
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

func (a *Application) htmlTemplateEn(w http.ResponseWriter, r *http.Request) {
	a.htmlTemplate(w, r, a.getFeedTitles(2, 15))
}

func (a *Application) htmlTemplateFi(w http.ResponseWriter, r *http.Request) {
	a.htmlTemplate(w, r, a.getFeedTitles(1, 15))
}

func (a *Application) htmlTemplate(w http.ResponseWriter, r *http.Request, result Result) {

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
		t.Execute(gw, result)
		gw.Close()
	} else {
		t.Execute(w, result)
	}
}

func (a *Application) addCategoryShowNamesAndMetaData(items []rss.Item, language int) Result {
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
	result := Result{}
	result.Items = a.AddCategoryEnNames(items)
	result.Description = a.addDescription(language)
	result.Lang = language
	return result
}

func (a *Application) addDescription(language int) string {
	if language == 1 {
		return "Uusimmat uutiset yhdestä lähteestä - www.uutispuro.fi"
	} else if language == 2 {
		return "News titles from one source - www.uutispuro.fi"
	}
	return "News titles from one source - www.uutispuro.fi"
}

func (a *Application) AddCategoryEnNames(items []rss.Item) []rss.Item {
	for i := range items {
		cat := items[i].Category.Name
		if cat == "IT ja media" {
			items[i].Category.EnName = "Digital media"
		} else if cat == "Digi" {
			items[i].Category.EnName = "Digital media"
		} else if cat == "TV ja elokuvat" {
			items[i].Category.EnName = "TV and movies"
		} else if cat == "Asuminen" {
			items[i].Category.EnName = "Home and living"
		} else if cat == "Kotimaa" {
			items[i].Category.EnName = "Domestic"
		} else if cat == "Kulttuuri" {
			items[i].Category.EnName = "Culture"
		} else if cat == "Matkustus" {
			items[i].Category.EnName = "Travel"
		} else if cat == "Pelit" {
			items[i].Category.EnName = "Games"
		} else if cat == "Ruoka" {
			items[i].Category.EnName = "Food"
		} else if cat == "Talous" {
			items[i].Category.EnName = "Economy"
		} else if cat == "Terveys" {
			items[i].Category.EnName = "Health"
		} else if cat == "Tiede" {
			items[i].Category.EnName = "Science"
		} else if cat == "Ulkomaat" {
			items[i].Category.EnName = "Foreign"
		} else if cat == "Urheilu" {
			items[i].Category.EnName = "Sports"
		} else if cat == "Viihde" {
			items[i].Category.EnName = "Entertainment"
		} else if cat == "Blogit" {
			items[i].Category.EnName = "Blogs"
		} else if cat == "Naiset ja muoti" {
			items[i].Category.EnName = "Women and fashion"
		}
	}
	return items
}

func (a *Application) getPage(req *http.Request) int {
	r, err := regexp.Compile("[0-9]+")
	if err != nil {
		return 0
	}
	res := r.FindString(req.URL.Path)
	page, err2 := strconv.Atoi(res)
	if err2 != nil {
		return 0
	}
	return page
}

func (a *Application) parseQueryValues(req *http.Request, value string) int {
	vals := req.URL.Query()
	val := vals[value]
	if val != nil {
		v, err := strconv.ParseInt(val[0], 10, 0)
		if err != nil {
			log.Println("page parsing error")
			return 0
		}
		return int(v)
	}
	return 0
}
