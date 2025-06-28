package session

import (
	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte("super-secret-key"))

func GetStore() *sessions.CookieStore {
	return store
}