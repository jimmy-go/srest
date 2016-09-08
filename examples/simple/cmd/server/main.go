// Package main contains full working example for srest.
//		EXAMPLE PROJECT
// Task list (To-Do) with i18n, i10n and user sessions
// controlled by sqlite.
//
// The MIT License (MIT)
//
// Copyright (c) 2016 Angel Del Castillo
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package main

import (
	"flag"
	"log"
	"net/http"
	"runtime"

	"github.com/jimmy-go/srest"
	"github.com/jimmy-go/srest/examples/simple/api/friends"
	"github.com/jimmy-go/srest/examples/simple/controllers/home"
	"github.com/jimmy-go/srest/examples/simple/dai"
)

var (
	port       = flag.Int("port", 0, "Listen port")
	connectURL = flag.String("db", "", "PostgreSQL connection url.")
	views      = flag.String("templates", "", "Templates files dir.")
	static     = flag.String("static", "", "Static dir.")
)

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())
	log.SetFlags(0)
	log.Printf("templates dir [%v]", *views)
	log.Printf("static dir [%v]", *static)
	log.SetFlags(log.Lshortfile)

	// connect to database. Mock database.
	err := dai.Connect(*connectURL)
	if err != nil {
		log.Fatal(err)
	}

	// load template views only for this project
	err = srest.LoadViews(*views, srest.DefaultFuncMap)
	if err != nil {
		log.Fatal(err)
	}

	m := srest.New(nil)
	m.Get("/static", srest.Static("/static", *static))
	m.Use("/v1/api/friends", friends.New())
	m.Use("/v1/api/mid", friends.New(), mwOne, mwTwo)
	m.Use("/home", &home.API{})
	<-m.Run(*port)
	log.Printf("Closing database connections")
	dai.Close()
	log.Printf("Done")
}

func mwOne(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		prev := r.URL.Query().Get("stop")
		if prev == "one" {
			w.Write([]byte("skipped from one"))
			return
		}
		h.ServeHTTP(w, r)
	})
}

func mwTwo(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		prev := r.URL.Query().Get("stop")
		if prev == "two" {
			w.Write([]byte("skipped from two"))
			return
		}
		h.ServeHTTP(w, r)
	})
}
