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
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"
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
	pool     chan *http.Client
	host     string
	targetsc chan *Target
	duration time.Duration
	done     chan struct{}
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
		pool:     ccs,
		host:     host,
		targetsc: make(chan *Target, users),
		duration: duration,
		done:     make(chan struct{}, 1),
	}
	return a
}

// Hit register endpoint for attack with model schema.
func (h *Attacker) Hit(path string, model interface{}) {
	// test only list
	params := url.Values{}
	h.targetsc <- &Target{
		URL:    h.host + path,
		Method: "GET",
		Data:   params,
	}
	return
	h.targetsc <- &Target{
		URL:    h.host + path + "/:id",
		Method: "GET",
		Data:   params,
	}
	h.targetsc <- &Target{
		URL:    h.host + path,
		Method: "POST",
		Data:   params,
	}
	h.targetsc <- &Target{
		URL:    h.host + path,
		Method: "PUT",
		Data:   params,
	}
	h.targetsc <- &Target{
		URL:    h.host + path + "/:id",
		Method: "DELETE",
		Data:   params,
	}
}

// Run func
func (h *Attacker) Run() chan struct{} {
	go func() {
		for {
			select {
			case tar := <-h.targetsc:
				h.targetsc <- tar
				if len(h.done) > 0 {
					log.Printf("Run : return closed chan targets")
					return
				}
				req, err := http.NewRequest(tar.Method, tar.URL, bytes.NewBufferString(tar.Data.Encode()))
				if err != nil {
					log.Printf("Run : err [%s]", err)
					continue
				}
				client, ok := <-h.pool
				if !ok {
					log.Printf("Run : return closed chan pool")
					return
				}
				res, err := client.Do(req)
				h.pool <- client
				if err != nil {
					log.Printf("Run : err [%s]", err)
					continue
				}
				// log.Printf("Run : status code [%v] url [%s]", res.StatusCode, tar.URL)
				if res.StatusCode != http.StatusOK {
					b, _ := ioutil.ReadAll(res.Body)
					log.Printf("Run : err status code [%v] method [%s] url [%s] body [%v]", res.StatusCode, tar.Method, tar.URL, string(b))
				}
				res.Body.Close()
			}
		}
	}()
	go func() {
		<-time.After(h.duration)
		h.Stop()
	}()
	return h.done
}

// Stop finishes all attackers routines.
func (h *Attacker) Stop() {
	select {
	case h.done <- struct{}{}:
	default:
	}
}

var (
	errs = make(map[string]bool)
	mut  sync.RWMutex
)

func logError(err error) {
	log.Printf("Error : [%s]", err)
}
