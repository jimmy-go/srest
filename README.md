# Simple RESTful toolkit

[![License MIT](https://img.shields.io/npm/l/express.svg)](http://opensource.org/licenses/MIT)
[![Build Status](https://travis-ci.org/jimmy-go/srest.svg?branch=master)](https://travis-ci.org/jimmy-go/srest)
[![Go Report Card](https://goreportcard.com/badge/github.com/jimmy-go/srest)](https://goreportcard.com/report/github.com/jimmy-go/srest)
[![GoDoc](http://godoc.org/github.com/jimmy-go/srest?status.png)](http://godoc.org/github.com/jimmy-go/srest)
[![Coverage Status](https://coveralls.io/repos/github/jimmy-go/srest/badge.svg?branch=master)](https://coveralls.io/github/jimmy-go/srest?branch=master)

srest goal it's help you building sites and RESTful APIs servers with clear
code and fast execution, without enslave you to complicated frameworks rules.
Uses a thin layer over another already useful toolkits:

[bmizerany/pat](https://github.com/bmizerany/pat)

[gorilla/schema](https://github.com/gorilla/schema)

----

Current version is under 1.0 so some breaking changes can be allowed.

Installation:
```
go get github.com/jimmy-go/srest
```

Usage:
```
    m := srest.New(nil) // init a new srest without TLS configuration.
	m.Get("/static", srest.Static("/static", *static)) // declare a static dir.
    m.Use("/v1/api/friends", friends.New()) // satisfies RESTfuler. This generates GET GET/:id POST PUT and DELETE/:id endpoints
    m.Use("/home", &home.API{}) // satisfies RESTfuler. This generates GET GET/:id POST PUT and DELETE/:id endpoints
    m.Get("/custom", myHTTPHandlerFunc) // you can access all pat methods directly too.
    <-m.Run(55555) // Run start a server on port 55555, if TLS support is needed take a look on srest.Options.
```

You need a RESTfuler interface and for your models Modeler interface.

```
type RESTfuler interface {
	Create(w http.ResponseWriter, r *http.Request)
	One(w http.ResponseWriter, r *http.Request)
	List(w http.ResponseWriter, r *http.Request)
	Update(w http.ResponseWriter, r *http.Request)
	Delete(w http.ResponseWriter, r *http.Request)
}

type Modeler interface {
	IsValid() error
}
```

You can pass middlewares too:
```
    m.Use("/v1/api/friends", friends.New(), Mid1, Mid2, Mid3 )
    m.Get("/customhandler", func(w http.ResponseWriter,r *http.Request){}, Mid1, Mid2, Mid3)
```

Example:
```
package users

// User model implements Modeler interface
type User struct {
	Name  string `db:"name" json:"name"`
	Email string `db:"email" json:"email"`
}

func (u *User) IsValid() bool {
    // do validation here
	return true
}

// API struct implements RESTfuler interface
type API struct{}

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

func (a *API) One(w http.ResponseWriter, r *http.Request) {
	srest.JSON(w, "some response")
}

func (a *API) List(w http.ResponseWriter, r *http.Request) {
	srest.JSON(w, "some list")
}

// We don't use this but is needed for RESTfuler interface
func (a *API) Update(w http.ResponseWriter, r *http.Request) {}

// We don't use this but is needed for RESTfuler interface
func (a *API) Delete(w http.ResponseWriter, r *http.Request) {}
```

###### breaking changes:

* replace ```srest.Static("/public", "mydir")``` for ```srest.Get("/public/", srest.Static("/public/", "mydir"))```

#### ToDo:

* Benchmark for Render. If needed implement Render with templates pool.
* Add support for status 503.
* Change example database to sqlite.
* Complete module stress and example stress.
* You can make stress tests using the package srest/stress with only a Modeler interface (described below).

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
