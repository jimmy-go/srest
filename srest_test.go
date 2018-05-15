package srest

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAutoSort(t *testing.T) {
	done := make(chan struct{}, 1)
	defer func() {
		err := recover()
		assert.EqualValues(t, nil, err)
	}()
	go func() {
		defer func() {
			err := recover()
			assert.EqualValues(t, nil, err)
		}()
		m := New(nil)
		m.Get("/", say("GET root"))
		m.Get("/me", say("GET home"))
		m.Get("/me/:id", say("GET me detail"))
		m.Get("/me/:id/name", say("GET me name"))

		m.Post("/", say("POST root"))
		m.Post("/me", say("POST home"))
		m.Post("/me/:id", say("POST me detail"))
		m.Post("/me/:id/name", say("POST me name"))

		m.Put("/", say("PUT root"))
		m.Put("/me", say("PUT home"))
		m.Put("/me/:id", say("PUT me detail"))
		m.Put("/me/:id/name", say("PUT me name"))

		m.Del("/", say("DELETE root"))
		m.Del("/me", say("DELETE home"))
		m.Del("/me/:id", say("DELETE me detail"))
		m.Del("/me/:id/name", say("DELETE me name"))

		m.Run(9000)
		<-done
	}()
	<-time.After(time.Second)
	table := []struct {
		Purpose, Method, URL, Exp string
	}{
		{"1. OK: get root", "GET", "http://localhost:9000/", "GET root"},
		{"2. OK: get home", "GET", "http://localhost:9000/me", "GET home"},
		{"3. OK: get home/", "GET", "http://localhost:9000/me/", "GET home"},
		{"4. OK: get detail", "GET", "http://localhost:9000/me/2", "GET me detail-%3Aid=2"},
		{"5. OK: get name", "GET", "http://localhost:9000/me/2/name", "GET me name-%3Aid=2"},

		{"6. OK: post root", "POST", "http://localhost:9000/", "POST root"},
		{"7. OK: post home", "POST", "http://localhost:9000/me", "POST home"},
		{"8. OK: post home/", "POST", "http://localhost:9000/me/", "POST home"},
		{"9. OK: post detail", "POST", "http://localhost:9000/me/2", "POST me detail-%3Aid=2"},
		{"10. OK: post name", "POST", "http://localhost:9000/me/2/name", "POST me name-%3Aid=2"},

		{"11. OK: put root", "PUT", "http://localhost:9000/", "PUT root"},
		{"12. OK: put home", "PUT", "http://localhost:9000/me", "PUT home"},
		{"13. OK: put home/", "PUT", "http://localhost:9000/me/", "PUT home"},
		{"14. OK: put detail", "PUT", "http://localhost:9000/me/2", "PUT me detail-%3Aid=2"},
		{"15. OK: put name", "PUT", "http://localhost:9000/me/2/name", "PUT me name-%3Aid=2"},

		{"16. OK: delete root", "DELETE", "http://localhost:9000/", "DELETE root"},
		{"17. OK: delete home", "DELETE", "http://localhost:9000/me", "DELETE home"},
		{"18. OK: delete home/", "DELETE", "http://localhost:9000/me/", "DELETE home"},
		{"19. OK: delete detail", "DELETE", "http://localhost:9000/me/2", "DELETE me detail-%3Aid=2"},
		{"20. OK: delete name", "DELETE", "http://localhost:9000/me/2/name", "DELETE me name-%3Aid=2"},
	}
	for i := range table {
		x := table[i]
		client := &http.Client{
			Timeout: 100 * time.Millisecond,
		}
		actual, err := getBody(client, x.Method, x.URL)
		assert.Nil(t, err, x.Purpose)
		assert.EqualValues(t, x.Exp, actual, x.Purpose)
	}
	done <- struct{}{}
}

func say(message string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Encode()
		if q != "" {
			q = "-" + q
		}
		if _, err := fmt.Fprintln(w, message+q); err != nil {
			// Can't panic.
			panic(err)
		}
	})
}

func getBody(client *http.Client, method, uri string) (string, error) {
	req, err := http.NewRequest(method, uri, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if err := resp.Body.Close(); err != nil {
		return "", err
	}
	return string(b[:len(b)-1]), nil
}

func TestRouterMatch(t *testing.T) {
	done := make(chan struct{}, 1)
	go func() {
		m := New(nil)
		m.Get("/me", say("GET home"))
		m.Get("/me/:id", say("GET me detail"))
		m.Run(9001)
		<-done
	}()
	<-time.After(time.Second)
	table := []struct {
		Purpose, Method, URL, Exp string
	}{
		{"1. OK: get home", "GET", "http://localhost:9001/me", "GET home"},
		{"2. OK: get name", "GET", "http://localhost:9001/me/2", "GET me detail-%3Aid=2"},
		{"3. Fail: 404 for control", "GET", "http://localhost:9001/", "404 page not found"},
		{"4. Fail: Must 404", "GET", "http://localhost:9001/me/2/name", "404 page not found"},
	}
	for i := range table {
		x := table[i]
		client := &http.Client{
			Timeout: 100 * time.Millisecond,
		}
		actual, err := getBody(client, x.Method, x.URL)
		assert.Nil(t, err, x.Purpose)
		assert.EqualValues(t, x.Exp, actual, x.Purpose)
	}
	done <- struct{}{}
}
