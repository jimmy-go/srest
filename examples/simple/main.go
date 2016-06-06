package main

import (
	"flag"
	"log"

	"github.com/jimmy-go/srest"
	"github.com/jimmy-go/srest/examples/simple/api/friends"
	"github.com/jimmy-go/srest/examples/simple/controllers/home"
	"github.com/jimmy-go/srest/examples/simple/dai"
)

var (
	port   = flag.Int("port", 0, "Listen port")
	dbf    = flag.String("db", "", "Database connection url.")
	tmpls  = flag.String("templates", "", "Templates files dir.")
	static = flag.String("static", "", "Static dir.")
)

func main() {
	flag.Parse()
	log.SetFlags(0)
	log.Printf("templates dir [%v]", *tmpls)
	log.Printf("static dir [%v]", *static)

	// connect to database
	err := dai.Configure(&dai.Options{URL: *dbf, Workers: 10, Queue: 10})
	if err != nil {
		log.Fatal(err)
	}

	// load template views only for this project
	err = srest.LoadViews(*tmpls, true)
	if err != nil {
		log.Fatal(err)
	}

	m := srest.New(nil)
	m.Static("/static", *static)
	m.Use("/v1/api/friends", friends.New(""))
	m.Use("/home", &home.API{})
	<-m.Run(*port)
	log.Printf("Closing database connections")
	dai.Db.Close()
	log.Printf("Done")
}
