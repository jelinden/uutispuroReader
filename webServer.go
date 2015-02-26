package main

import (
	"bytes"
	"compress/gzip"
	//"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/jelinden/rssFetcher/rss"
	"github.com/jelinden/uutispuroReader/service"
	"github.com/jelinden/uutispuroReader/util"
	"gopkg.in/mgo.v2/bson"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var newsToBefetched = 150

type Application struct {
	Sessions   *service.Sessions
	CookieUtil *util.CookieUtil
}

func NewApplication() *Application {
	return &Application{}
}

func (a *Application) Init() {
	a.Sessions = service.NewSessions()
	a.CookieUtil = util.NewCookieUtil()
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
	r := gin.Default()
	r.LoadHTMLTemplates("templates/*")
	r.Static("/assets/", "./assets")

	r.GET("/", func(c *gin.Context) {
		lang, err := c.Request.Cookie("uutispuroLang")
		if err != nil {
			log.Println(err)
			c.Redirect(302, "/en/")
		} else if lang.Value == "fi" {
			c.Redirect(302, "/fi/")
		} else {
			c.Redirect(302, "/en/")
		}
	})
	r.GET("/fi/", func(c *gin.Context) {
		app.CookieUtil.SetCookie("uutispuroLang", "fi", c)
		obj := app.getFeedTitles(1, 15)
		c.HTML(200, "index.html", obj)
	})
	r.GET("/en/", func(c *gin.Context) {
		app.CookieUtil.SetCookie("uutispuroLang", "en", c)
		obj := app.getFeedTitles(2, 15)
		c.HTML(200, "index.html", obj)
	})
	r.GET("/fi/ws", func(c *gin.Context) {
		c.JSON(200, app.fetchRssItems(1))
	})
	r.GET("/en/ws", func(c *gin.Context) {
		c.JSON(200, app.fetchRssItems(2))
	})
	r.GET("/c/:id", func(c *gin.Context) {
		id := c.Params.ByName("id")
		app.saveClick(id)
	})
	r.GET("/l/:id", func(c *gin.Context) {
		id := c.Params.ByName("id")
		app.saveLike(id)
	})
	r.GET("/u/:id", func(c *gin.Context) {
		id := c.Params.ByName("id")
		app.saveUnlike(id)
	})
	r.GET("/category/:cat/:lang/ws", func(c *gin.Context) {
		lang := c.Params.ByName("lang")
		cat := c.Params.ByName("cat")
		if lang == "en" {
			c.JSON(200, app.fetchRssItemsByCategory(2, cat))
		} else {
			c.JSON(200, app.fetchRssItemsByCategory(1, cat))
		}
	})

	r.GET("/category/:cat/:lang", func(c *gin.Context) {
		lang := c.Params.ByName("lang")
		obj := service.Result{}
		cat := c.Params.ByName("cat")
		obj = app.getFeedCategoryTitles(app.langAsInt(lang), cat, 15)
		c.HTML(200, "index.html", obj)
	})

	fmt.Println("starting web server")
	r.Run(":9100")
}

func (a *Application) getFeedTitles(language int, limit int) service.Result {
	result := []rss.Item{}
	s := a.Sessions.Mongo.Clone()
	c := s.DB("uutispuro").C("rss")
	err := c.Find(bson.M{"language": language}).Sort("-date").Limit(limit).All(&result)
	if err != nil {
		fmt.Println("Fatal error " + err.Error())
	}
	return a.addCategoryShowNamesAndMetaData(result, language, "")
}

func (a *Application) getFeedCategoryTitles(language int, category string, limit int) service.Result {
	result := []rss.Item{}
	s := a.Sessions.Mongo.Clone()
	c := s.DB("uutispuro").C("rss")
	err := c.Find(
		bson.M{"language": language, "category.name": strings.ToUpper(category[0:1]) + category[1:]},
	).
		Sort("-date").
		Limit(limit).All(&result)
	if err != nil {
		fmt.Println("Fatal error " + err.Error())
	}
	return a.addCategoryShowNamesAndMetaData(result, language, category)
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

func (a *Application) fetchRssItems(lang int) map[string]interface{} {
	return map[string]interface{}{"d": a.getFeedTitles(lang, newsToBefetched)}
}

func (a *Application) fetchRssItemsByCategory(lang int, category string) map[string]interface{} {
	return map[string]interface{}{"d": a.getFeedCategoryTitles(lang, category, newsToBefetched)}
}

func (a *Application) rootHandler(w http.ResponseWriter, r *http.Request) {
	//pageNumber := a.getPage(r)
	//log.Println(r.RemoteAddr, r.RequestURI, pageNumber)
	attrs := strings.Split(r.URL.Path, "/")
	var content []byte = nil
	if len(attrs) > 3 {
		if strings.EqualFold(attrs[1], "fi") && strings.EqualFold(attrs[2], "category") {
			a.htmlCategoryTemplate(w, r, 1, attrs[3])
		} else if strings.EqualFold(attrs[1], "en") && strings.EqualFold(attrs[2], "category") {
			a.htmlCategoryTemplate(w, r, 2, attrs[3])
		}
	} else if strings.EqualFold(r.RequestURI, "/") {
		http.Redirect(w, r, "/en/", 301)
	} else if strings.HasPrefix(r.RequestURI, "/fi") || strings.HasPrefix(r.RequestURI, "/fi/") {
		a.htmlTemplateFi(w, r)
	} else if strings.HasPrefix(r.RequestURI, "/en") || strings.HasPrefix(r.RequestURI, "/en/") {
		a.htmlTemplateEn(w, r)
	} else if strings.HasSuffix(r.RequestURI, ".css") {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		w.Header().Set("Content-Encoding", "gzip")
		a.setHttpCacheHeaders(w.Header())
		content = a.openFileGzipped("styles" + r.RequestURI[strings.LastIndex(r.RequestURI, "/"):len(r.RequestURI)])
	} else if strings.HasSuffix(r.RequestURI, ".js") {
		w.Header().Set("Content-Type", "application/javascript")
		w.Header().Set("Content-Encoding", "gzip")
		a.setHttpCacheHeaders(w.Header())
		content = a.openFileGzipped("js" + r.RequestURI[strings.LastIndex(r.RequestURI, "/"):len(r.RequestURI)])
	} else if strings.HasSuffix(r.RequestURI, "/favicon.ico") {
		a.setHttpCacheHeaders(w.Header())
		content = a.openFileGzipped("img/favicon.ico")
	} else if strings.HasSuffix(r.RequestURI, ".png") {
		a.setHttpCacheHeaders(w.Header())
		w.Header().Set("Content-Type", "image/png")
		content = a.openFile("img" + r.RequestURI[strings.LastIndex(r.RequestURI, "/"):len(r.RequestURI)])
	} else if strings.HasSuffix(r.RequestURI, ".gif") {
		w.Header().Set("Content-Type", "image/gif")
		a.setHttpCacheHeaders(w.Header())
		content = a.openFile("img" + r.RequestURI[strings.LastIndex(r.RequestURI, "/"):len(r.RequestURI)])
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
	fmt.Fprintf(w, "%s", content)
}

func (a *Application) setHttpCacheHeaders(header http.Header) {
	header.Set("Cache-Control", fmt.Sprintf("max-age=%d, public, must-revalidate, proxy-revalidate", 60*60*24*7*4))
	header.Set("Last-Modified", time.Now().Format(http.TimeFormat))
	header.Set("Expires", time.Now().AddDate(60, 0, 0).Format(http.TimeFormat))
}

func (a *Application) openFile(fileName string) []byte {
	fmt.Println(fileName)
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
	a.htmlTemplate(w, r, a.getFeedTitles(2, 15), "html/index.html")
}

func (a *Application) htmlTemplateFi(w http.ResponseWriter, r *http.Request) {
	a.htmlTemplate(w, r, a.getFeedTitles(1, 15), "html/index.html")
}

func (a *Application) htmlCategoryTemplate(w http.ResponseWriter, r *http.Request, lang int, category string) {
	a.htmlTemplate(w, r, a.getFeedCategoryTitles(lang, category, 15), "html/index.html")
}

func (a *Application) htmlTemplate(w http.ResponseWriter, r *http.Request, result service.Result, tString string) {
	t, err := template.ParseFiles(tString)
	if err != nil {
		log.Printf("Template gave: %s", err)
	}
	cangzip := strings.Index(r.Header.Get("Accept-Encoding"), "gzip") > -1
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Expires", "-1")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
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

func (a *Application) addCategoryShowNamesAndMetaData(items []rss.Item, language int, category string) service.Result {
	for i := range items {
		items[i].Category.StyleName = items[i].Category.Name
	}
	result := service.Result{}
	result.Items = a.AddCategoryEnNames(items)
	result.Description = a.addDescription(language, category)
	result.PageTitle = a.addPageTitle(language, category)
	result.Lang = language
	return result
}

func (a *Application) addDescription(language int, category string) string {
	if category == "" {
		if language == 1 {
			return "Uusimmat uutiset yhdestä lähteestä - www.uutispuro.fi"
		} else if language == 2 {
			return "News titles from one source - www.uutispuro.fi"
		}
		return "News titles from one source - www.uutispuro.fi"
	} else {
		if language == 1 {
			return category + " - www.uutispuro.fi"
		} else if language == 2 {
			return a.getEnCategoryName(category) + " - www.uutispuro.fi"
		}
		return a.getEnCategoryName(category) + " - www.uutispuro.fi"
	}
}

func (a *Application) addPageTitle(language int, category string) string {
	if category == "" {
		if language == 1 {
			return "Uusimmat uutiset - Uutispuro"
		} else if language == 2 {
			return "News titles - Uutispuro"
		}
		return "News titles - Uutispuro"
	} else {
		if language == 1 {
			return category + " - Uutispuro"
		} else if language == 2 {
			return a.getEnCategoryName(category) + " - Uutispuro"
		}
		return a.getEnCategoryName(category) + " - Uutispuro"
	}
}

func (a *Application) AddCategoryEnNames(items []rss.Item) []rss.Item {
	for i := range items {
		cat := items[i].Category.Name
		items[i].Category.EnName = a.getEnCategoryName(cat)
	}
	return items
}

func (a *Application) getEnCategoryName(cat string) string {
	if cat == "Digi" {
		return "Tech"
	} else if cat == "Elokuvat" {
		return "Movies"
	} else if cat == "Koti" {
		return "Home"
	} else if cat == "Kotimaa" {
		return "Domestic"
	} else if cat == "Kulttuuri" {
		return "Culture"
	} else if cat == "Matkustus" {
		return "Travel"
	} else if cat == "Pelit" {
		return "Games"
	} else if cat == "Ruoka" {
		return "Food"
	} else if cat == "Talous" {
		return "Economy"
	} else if cat == "Terveys" {
		return "Health"
	} else if cat == "Tiede" {
		return "Science"
	} else if cat == "Ulkomaat" {
		return "Foreign"
	} else if cat == "Urheilu" {
		return "Sports"
	} else if cat == "Viihde" {
		return "Entertainment"
	} else if cat == "Blogit" {
		return "Blogs"
	} else if cat == "Naiset" {
		return "Women"
	} else {
		return ""
	}
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

func (a *Application) langAsInt(lang string) int {
	if lang == "en" {
		return 2
	} else {
		return 1
	}
}
