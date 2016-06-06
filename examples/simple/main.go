package main

import (
	"flag"
	"log"

	"github.com/jimmy-go/srest"
	"github.com/jimmy-go/srest/examples/simple/dai"
	"github.com/jimmy-go/srest/examples/simple/friends"
	"github.com/jimmy-go/srest/views"
)

// TODO; change path to views.
var (
	port  = flag.Int("port", 0, "Listen port")
	dbf   = flag.String("db", "", "Database connection url.")
	tmpls = flag.String("templates", "", "Templates files dir.")
)

func main() {
	flag.Parse()
	log.SetFlags(0)
	log.Printf("templates dir [%v]", *tmpls)
	log.Printf("db [%v]", *dbf)

	// connect to database
	err := dai.Configure(&dai.Options{URL: *dbf, Workers: 10, Queue: 10})
	if err != nil {
		log.Fatal(err)
	}

	// load template views only for this project
	err = views.Configure(*tmpls)
	if err != nil {
		log.Fatal(err)
	}

	m := srest.New(nil)
	m.Use("/v1/api/friends", friends.New(""))
	<-m.Run(*port)
	log.Printf("Closing database connections")
	dai.Db.Close()
	log.Printf("Done")
}
