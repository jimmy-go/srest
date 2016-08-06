// Package main contains full working example for srest.
// EXAMPLE PROJECT
// Task list (To-Do) with i18n, i10n and user sessions
// controlled by sqlite.
package main

import (
	"flag"
	"log"
	"net/http"
	"runtime"
	"time"

	"github.com/jimmy-go/srest"
	"github.com/jimmy-go/srest/examples/simple/api/friends"
	"github.com/jimmy-go/srest/examples/simple/controllers/home"
	"github.com/jimmy-go/srest/examples/simple/dai"
)

var (
	port    = flag.Int("port", 0, "Listen port")
	dbf     = flag.String("db", "", "Database connection url.")
	views   = flag.String("templates", "", "Templates files dir.")
	static  = flag.String("static", "", "Static dir.")
	workers = flag.Int("workers", 1, "Worker pool size.")
	queue   = flag.Int("queue", 10, "Queue length.")
)

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())
	log.SetFlags(0)
	log.Printf("templates dir [%v]", *views)
	log.Printf("static dir [%v]", *static)

	// connect to database
	conf := &dai.Options{
		URL:     *dbf,
		Workers: *workers,
		Queue:   *queue,
	}
	now := time.Now()
	err := dai.Connect(conf)
	log.Printf("Database connection time: [%s]", time.Since(now))
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
	dai.Db.Close()
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
