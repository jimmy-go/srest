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
)

// Target struct
type Target struct {
	URL    string
	Method string
	Data   url.Values
}

// Attacker struct.
type Attacker struct {
	pool    chan *http.Client
	host    string
	targets []*Target
	Done    chan struct{}
}

// New returns a new attacker.
func New(host string, users int, duration time.Duration) *Attacker {
	// populate http.Clients pool
	ccs := make(chan *http.Client, users)
	for i := 0; i < users; i++ {
		cl := &http.Client{
			Timeout: 60 * time.Second,
		}
		ccs <- cl
	}

	a := &Attacker{
		pool: ccs,
		host: host,
		Done: make(chan struct{}, 1),
	}
	return a
}

// Hit attacks endpoint.
func (h *Attacker) Hit(path string, model interface{}) {
	h.targets = append(h.targets, &Target{
		URL:    h.host + path + "/:id",
		Method: "GET",
	})
	h.targets = append(h.targets, &Target{
		URL:    h.host + path,
		Method: "GET",
	})
	h.targets = append(h.targets, &Target{
		URL:    h.host + path,
		Method: "POST",
	})
	h.targets = append(h.targets, &Target{
		URL:    h.host + path,
		Method: "PUT",
	})
	h.targets = append(h.targets, &Target{
		URL:    h.host + path + "/:id",
		Method: "DELETE",
	})
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
				client := <-h.pool
				uri := h.targets[0].URL
				_, err := client.Get(uri)
				h.pool <- client
				if err != nil {
					logError(err)
				}
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
