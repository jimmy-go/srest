// Package stress allows fast creation of load testing environments
/*

TODO; add example usage

*/
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
package stress

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/jimmy-go/jobq"
	"github.com/jimmy-go/srest"
)

// Target struct
type Target struct {
	URL    string
	Method string
	Data   url.Values
}

// Attacker struct.
type Attacker struct {
	pool     *jobq.Dispatcher
	host     string
	clientsc chan *http.Client
	targets  []*Target
	Done     chan struct{}
}

// New returns a new attacker.
func New(host string, users int, d time.Duration) *Attacker {
	// init worker pool
	jq, err := jobq.New(users, users)
	if err != nil {
		panic(err)
	}

	// populate http.Clients pool
	cclients := make(chan *http.Client, users)
	for i := 0; i < users; i++ {
		cl := &http.Client{
			Timeout: 60 * time.Second,
		}
		cclients <- cl
	}

	a := &Attacker{
		pool:     jq,
		clientsc: cclients,
		host:     host,
		Done:     make(chan struct{}, 1),
	}
	return a
}

// Hit attacks endpoint.
func (h *Attacker) Hit(path string, API srest.RESTfuler, model interface{}) {
	h.targets = append(h.targets, &Target{
		URL:    h.host + path,
		Method: "TODO",
	})
}

// HitStatic attacks endpoint.
func (h *Attacker) HitStatic(path, dir string) {
}

// Run func
func (h *Attacker) Run() chan os.Signal {
	go func() {
		for {
			select {
			case <-h.Done:
				log.Printf("Exit")
				return
			// case <-time.After(5 * time.Millisecond):
			default:
				h.pool.Add(func() error {
					client := <-h.clientsc
					uri := h.targets[0].URL
					_, err := client.Get(uri)
					h.clientsc <- client
					if err != nil {
						logError(err)
						return err
					}
					// log.Printf("Hit : Status [%s]", res.Status)
					return nil
				})
			}
		}
	}()
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	return c
}

var (
	errs = make(map[string]bool)
	mut  sync.RWMutex
)

func logError(err error) {
	log.Printf("Error : [%s]", err)
	return
	//	mut.RLock()
	//	defer mut.RUnlock()
	//	_, ok := errs[err.Error()]
	//	if ok {
	//		return
	//	}
	//	errs[err.Error()] = true
}
