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
    // declare a new srest without TLS configuration.
    m := srest.New(nil)

    // static server endpoint.
	m.Get("/static", srest.Static("/static", *static))

    // friends.New() return a struct that satisfies RESTfuler.
    // This generates endpoints:
    // /v1/api/friends GET
    // /v1/api/friends/:id GET
    // /v1/api/friends POST
    // /v1/api/friends/:id PUT
    // /v1/api/friends/:id DELETE
    m.Use("/v1/api/friends", friends.New())

    // for custom endpoints you can use .Get .Post .Put and .Del
    // you can pass middlewares too.
    m.Get("/custom", myHTTPHandler, Mid1, Mid2, Mid3)

    // you can access mux directly too.
    // (you can't add middlewares so easily with this way.)
    m.Mux.Post("/me", myHTTPHandlerFunc)

    // Run call http.ListenAndServe or ListenAndServeTLS
    // (view srest.Options for TLS config)
    // until SIGTERM or SIGINT signal.
    <-m.Run(55555)

    // when you call Use Method in srest a RESTfuler interface is required.
    type RESTfuler interface {
        Create(w http.ResponseWriter, r *http.Request)
        One(w http.ResponseWriter, r *http.Request)
        List(w http.ResponseWriter, r *http.Request)
        Update(w http.ResponseWriter, r *http.Request)
        Delete(w http.ResponseWriter, r *http.Request)
    }
```

You need an easy way for params validation? take a look at Modeler interface
type Modeler interface {
	IsValid() error
}

example:
```
type Params struct{
    Name string `schema:"name"`
    LastName string `schema:"last_name"`
}

// my model validation
func(m *Params) IsValid() error{
    if len(m.Name) < 1 {
        return errors.New("model: param name is required")
    }
    return nil
}

var p Params
// Bind binds url.Values to struct using gorilla schema
err := srest.Bind(req.PostForm, &p)
check errors...
```

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
