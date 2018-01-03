package srest

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAutoSort(t *testing.T) {
	defer func() {
		err := recover()
		assert.EqualValues(t, nil, err)
	}()
	go func() {
		m := New(nil)
		m.Get("/me", say("GET home"))
		m.Get("/me/:id", say("GET me detail"))
		m.Get("/me/:id/name", say("GET me name"))
		<-m.Run(9000)
	}()

	table := []struct {
		Purpose, URL, Exp string
	}{
		{"1. OK: get home", "http://localhost:9000/me", "GET home"},
		{"2. OK: get home", "http://localhost:9000/me/", "GET home"},
		{"3. OK: get detail", "http://localhost:9000/me/2", "GET me detail"},
		{"4. OK: get name", "http://localhost:9000/me/2/name", "GET me name"},
	}
	for _, x := range table {
		actual := getBody(x.URL)
		assert.EqualValues(t, x.Exp, actual, x.Purpose)
	}
}

func say(message string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprintln(w, message); err != nil {
			panic(err)
		}
	})
}

func getBody(uri string) string {
	resp, err := http.Get(uri)
	if err != nil {
		panic(err)
	}
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	return string(b[:len(b)-1])
}
