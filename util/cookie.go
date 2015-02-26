package util

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type CookieUtil struct {
}

func NewCookieUtil() *CookieUtil {
	return &CookieUtil{}
}

func (c *CookieUtil) SetCookie(name string, value string, context *gin.Context) {
	expire := time.Now().AddDate(1, 0, 1)
	cookie := http.Cookie{name, value, "/", ".uutispuro.fi", expire, expire.Format(time.UnixDate), 41472000, false, false, name + "=" + value, []string{name + "=" + value}}
	http.SetCookie(context.Writer, &cookie)
}
