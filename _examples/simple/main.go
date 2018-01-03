package main

import (
	"flag"
	"log"
	"net/http"
	"runtime"

	"github.com/jimmy-go/srest"
)

var (
	port = flag.Int("port", 9000, "Listen port")
)

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())
	m := srest.New(nil)
	m.Get("/", http.HandlerFunc(homeHandler))
	<-m.Run(*port)
	log.Println("Done")
}

// Response type for common API responses.
type Response struct {
	Message string `json:"response"`
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	res := &Response{"hello world"}
	if err := srest.JSON(w, res); err != nil {
		log.Printf("home : json : err [%s]", err)
	}
}
