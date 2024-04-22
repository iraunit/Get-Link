package adapter

import (
	"github.com/iraunit/get-link-backend/util"
	"net/http"
	"time"
)

func SetCookie(projectType string, tokenString string, domain string) *http.Cookie {
	cookie := &http.Cookie{
		Name:    "Authorization",
		Value:   tokenString,
		Expires: time.Now().Add(time.Hour * 24),
		Path:    "/",
	}
	if projectType == util.PRODUCTION {
		cookie.Secure = true
		cookie.HttpOnly = false
		cookie.Domain = domain
		cookie.SameSite = http.SameSiteStrictMode
	}
	return cookie
}
