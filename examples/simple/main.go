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
	tmpls   = flag.String("templates", "", "Templates files dir.")
	static  = flag.String("static", "", "Static dir.")
	workers = flag.Int("workers", 1, "Worker pool size.")
	queue   = flag.Int("queue", 10, "Queue length.")
)

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())
	log.SetFlags(0)
	log.Printf("templates dir [%v]", *tmpls)
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
	err = srest.LoadViews(*tmpls)
	if err != nil {
		log.Fatal(err)
	}

	m := srest.New(nil)
	m.Static("/static", *static)
	m.Use("/v1/api/friends", friends.New())
	m.Use("/v1/api/mid", friends.New(), mwOne(), mwTwo())
	m.Use("/home", &home.API{})
	<-m.Run(*port)
	log.Printf("Closing database connections")
	dai.Db.Close()
	log.Printf("Done")
}

func mwOne() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("mwOne: [%v]", r.URL)
	})
}

func mwTwo() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("mwTwo: [%v]", r.URL)
		ff := r.URL.Query().Get("fail")
		if ff == "fail" {
			log.Printf("mwTwo: verification err")
			http.Error(w, "middleware Two says: bad request", http.StatusUnauthorized)
			return
		}
	})
}
