package services

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/pivotal-golang/conceal"
)

const (
	chunkSize   = 3500
	SessionName = "autoscaling"
)

type Session struct {
	cookieStore CookieStore
	values      map[string]string
}

func NewSession(encryptionKey []byte, sessionName string, request *http.Request, writer http.ResponseWriter) Session {
	store := NewCookieStore(request, writer, encryptionKey, sessionName)
	values := store.Retrieve()

	return Session{
		cookieStore: store,
		values:      values,
	}
}

func (session Session) Get(key string) (string, bool) {
	value, ok := session.values[key]
	return value, ok
}

func (session Session) Set(key, value string) {
	session.values[key] = value
}

func (session Session) Values() map[string]string {
	return session.values
}

func (session Session) Reset() {
	for key, _ := range session.values {
		delete(session.values, key)
	}
}

func (session Session) Save() error {
	return session.cookieStore.Stash(session.values)
}

type CookieStore struct {
	request     *http.Request
	writer      http.ResponseWriter
	sessionName string
	cloak       conceal.Cloak
}

func NewCookieStore(request *http.Request, writer http.ResponseWriter, encryptionKey []byte, sessionName string) CookieStore {
	cloak, _ := conceal.NewCloak(encryptionKey)

	return CookieStore{
		request:     request,
		writer:      writer,
		sessionName: sessionName,
		cloak:       cloak,
	}
}

func (store CookieStore) Stash(values map[string]string) error {
	items, err := json.Marshal(values)
	if err != nil {
		return err
	}

	encryptedItems, err := store.cloak.Veil(items)
	if err != nil {
		return err
	}

	buffer := bytes.NewBufferString(string(encryptedItems))

	counter := 0
	for buffer.Len() > 0 {
		counter++
		chunk := buffer.Next(chunkSize)

		cookie := http.Cookie{
			Name:  store.sessionName + "-" + strconv.Itoa(counter),
			Value: string(chunk),
			Path:  "/",
		}

		http.SetCookie(store.writer, &cookie)
	}
	return nil
}

func (store CookieStore) Retrieve() map[string]string {
	counter := 0
	buffer := bytes.NewBuffer([]byte{})

	for {
		counter++
		cookie, err := store.request.Cookie(store.sessionName + "-" + strconv.Itoa(counter))

		if err != nil {
			if err == http.ErrNoCookie {
				break
			}
			return make(map[string]string)
		}
		_, err = buffer.WriteString(cookie.Value)
		if err != nil {
			return make(map[string]string)
		}
	}

	if buffer.Len() == 0 {
		return make(map[string]string)
	}

	decryptedValue, err := store.cloak.Unveil([]byte(buffer.String()))
	if err != nil {
		return make(map[string]string)
	}

	var values map[string]string
	err = json.Unmarshal([]byte(decryptedValue), &values)
	if err != nil {
		return make(map[string]string)
	}

	return values
}
