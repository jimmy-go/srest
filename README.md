#### Simple RESTful toolkit.

[![License MIT](https://img.shields.io/npm/l/express.svg)](http://opensource.org/licenses/MIT)
[![Build Status](https://travis-ci.org/jimmy-go/srest.svg?branch=master)](https://travis-ci.org/jimmy-go/srest)
[![Go Report Card](https://goreportcard.com/badge/github.com/jimmy-go/srest)](https://goreportcard.com/report/github.com/jimmy-go/srest)
[![GoDoc](http://godoc.org/github.com/jimmy-go/srest?status.png)](http://godoc.org/github.com/jimmy-go/srest)
[![Coverage Status](https://coveralls.io/repos/github/jimmy-go/srest/badge.svg?branch=master)](https://coveralls.io/github/jimmy-go/srest?branch=master)

Srest goal it's help you build sites and clear RESTful APIs webservices.
Without enslave you to complicated frameworks rules.
It's a thin layer over other useful toolkits:

[bmizerany/pat](https://github.com/bmizerany/pat)

[gorilla/schema](https://github.com/gorilla/schema)

#####Features:
* Endpoint declaration with middleware support.
* Input model validation.
* Templates made easy (and faster).
* Util for Fast to build Stress tests (still in development).

_Current version is under 1.0 some breaking changes can happen._

----

#####Install:
```
go get github.com/jimmy-go/srest
```

#####Usage:
```
    // declare a new srest without TLS configuration.
    m := srest.New(nil)

    // static server endpoint.
	m.Get("/public", srest.Static("/public/", PathToMyDir))

    // friends.New() return a struct that satisfies RESTfuler.
    // generates endpoints:
    // GET     /v1/api/friends
    // GET     /v1/api/friends/:id
    // POST    /v1/api/friends
    // PUT     /v1/api/friends/:id
    // DELETE  /v1/api/friends/:id
    m.Use("/v1/api/friends", friends.New())
    // with middlewares
    m.Use("/v1/api/friends", friends.New(), Mid1, Mid2, Mid3)

    // for custom endpoints you can use .Get .Post .Put
    // and .Del methods
    // you can pass middlewares too.
    m.Get("/custom", myHTTPHandler, Mid1, Mid2, Mid3)

    // you can access mux directly too.
    // (but you can't add middlewares easily this way.)
    m.Mux.Post("/me", myHTTPHandlerFunc)

    // Run call http.ListenAndServe or ListenAndServeTLS
    // (view srest.Options for TLS config)
    // until SIGTERM or SIGINT signal.
    <-m.Run(55555)

    // when you call Use Method in srest a RESTfuler interface
    // is required.
    type RESTfuler interface {
        Create(w http.ResponseWriter, r *http.Request)
        One(w http.ResponseWriter, r *http.Request)
        List(w http.ResponseWriter, r *http.Request)
        Update(w http.ResponseWriter, r *http.Request)
        Delete(w http.ResponseWriter, r *http.Request)
    }
```

You need an easy way for params validation? take a look at Modeler interface
```
type Modeler interface {
	IsValid() error
}
```

Example:
```
// my model
type Params struct{
    Name string `schema:"name"`
    LastName string `schema:"last_name"`
}

// my model validation
func(m *Params) IsValid() error {
    if len(m.Name) < 1 {
        return errors.New("model: name is required")
    }
    return nil
}

// some handlerfunc
func(w http.ResponseWriter, r *http.Request) {
    var p Params
    // Bind binds url.Values to struct using gorilla schema
    err := srest.Bind(r.PostForm, &p)
    check errors...
}
```

##### Working with html templates:
```
    // declare a new srest without TLS configuration.
    m := srest.New(nil)

    // load templates
    err := srest.LoadViews(PathToDir, srest.DefaultFuncMap)
    check errors...

    // start server as normal
    <-m.Run(7070)

    // some http handler func.
    func(w http.ResponseWriter, r *http.Request) {
        v := map[string]interface{}{"some":"A"}
        err := srest.Render(w, "home.html", v)
        check errors...
    }
```

Take a look at the working example with all features on examples dir.

#####ToDo:

* Benchmark for Render. If needed implement Render with templates pool.
* Add support for status 503.
* Change example database to sqlite.
* Complete module stress and example stress.
* Make stress tests using the package srest/stress.

#####License:

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
