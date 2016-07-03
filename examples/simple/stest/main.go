package main

import (
	"flag"
	"log"
	"runtime"
	"time"

	"github.com/jimmy-go/srest/examples/simple/api/friends"
	"github.com/jimmy-go/srest/examples/simple/controllers/home"
	"github.com/jimmy-go/srest/stress"
)

var (
	static = flag.String("static", "", "Static dir.")
	host   = flag.String("host", "", "Destination host for stress.")
	users  = flag.Int("users", 60, "Users count.")
)

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())
	log.SetFlags(log.Lshortfile | log.Lmicroseconds)
	log.Printf("workers [%v]", *users)
	s := stress.New(*host, *users, 60*time.Second)
	s.HitStatic("/static", *static)
	s.Hit("/v1/api/friends", friends.New(), &friends.Friend{})
	s.Hit("/home", &home.API{}, &home.Home{})
	<-s.Run()
}
