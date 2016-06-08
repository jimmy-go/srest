# Simple RESTful toolkit

[![License MIT](https://img.shields.io/npm/l/express.svg)](http://opensource.org/licenses/MIT)
[![Build Status](https://travis-ci.org/jimmy-go/srest.svg?branch=master)](https://travis-ci.org/jimmy-go/srest)
[![Go Report Card](https://goreportcard.com/badge/github.com/jimmy-go/srest)](https://goreportcard.com/report/github.com/jimmy-go/srest)
[![GoDoc](http://godoc.org/github.com/jimmy-go/srest?status.png)](http://godoc.org/github.com/jimmy-go/srest)
[![Coverage Status](https://coveralls.io/repos/github/jimmy-go/srest/badge.svg?branch=master&1)](https://coveralls.io/github/jimmy-go/srest?branch=master)

----

Usage:
```
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

	// load template views
	err = srest.LoadViews(*tmpls)
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
```

You need a RESTfuler interface and for your models Modeler interface.

```
package users

// User model
type User struct {
	Name  string `db:"name" json:"name"`
	Email string `db:"email" json:"email"`
}

// IsValid satisfies modeler interface.
func (u *User) IsValid() bool {
    // do validation here
	return true
}

// API struct
type API struct{}

// Create func
func (a *API) Create(w http.ResponseWriter, r *http.Request) {
	var m *Friend
	err := srest.Bind(r, &m)
	if err != nil {
		srest.JSON(w, err)
		return
	}
    // Logic here
	srest.JSON(w, "some response")
}

// One func
func (a *API) One(w http.ResponseWriter, r *http.Request) {
	srest.JSON(w, "some response")
}

// List func
func (a *API) List(w http.ResponseWriter, r *http.Request) {
	srest.JSON(w, "some response")
}

// Update func. We don't use this but is needed for RESTfuler interface
func (a *API) Update(w http.ResponseWriter, r *http.Request) {}

// Delete func. We don't use this but is needed for RESTfuler interface
func (a *API) Delete(w http.ResponseWriter, r *http.Request) {}
```

ToDo

* Stress util


##### License

The MIT License (MIT)

Copyright (c) 2016 Angel Del Castillo

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
