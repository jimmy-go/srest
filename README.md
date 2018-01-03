## RESTful toolkit.

[![License MIT](https://img.shields.io/npm/l/express.svg)](http://opensource.org/licenses/MIT)
[![Build Status](https://travis-ci.org/jimmy-go/srest.svg?branch=master)](https://travis-ci.org/jimmy-go/srest)
[![Go Report Card](https://goreportcard.com/badge/github.com/jimmy-go/srest)](https://goreportcard.com/report/github.com/jimmy-go/srest)
[![GoDoc](http://godoc.org/github.com/jimmy-go/srest?status.png)](http://godoc.org/github.com/jimmy-go/srest)
[![Coverage Status](https://coveralls.io/repos/github/jimmy-go/srest/badge.svg?branch=master)](https://coveralls.io/github/jimmy-go/srest?branch=master)

Srest goal it's help you build sites and clear RESTful services without enslave
you to complicated frameworks rules.
It's a thin layer over other useful toolkits:

[bmizerany/pat](https://github.com/bmizerany/pat)

[gorilla/schema](https://github.com/gorilla/schema)

### Features:

* Endpoint declaration and validation with middleware support.
* Payload validation.
* Templates made easy.
* Simple.

### Install:
```
go get gopkg.in/jimmy-go/srest.v0
```

### Usage:

#### Simplest example:

```
// Declare a new srest without TLS configuration. View srest.Options for TLS config.
m := srest.New(nil)

// Declare an endpoint.
m.Get("/hello", func(w http.ResponseWriter, r *http.Request) {})

// Run calls http.ListenAndServe or ListenAndServeTLS
// until SIGTERM or SIGINT signal is received.
<-m.Run(9000)
```

#### With middleware:

```
m := srest.New(nil)
m.Get("/hello", helloHandler, Mid1, Mid2, Mid3)

// Another more legible way to pass middleware:
c := []func(http.Handler) http.Handler{
    Mid1,
    Mid2,
    Mid3,
}
m.Get("/hello/2", someHandler, c...)
m.Get("/hello/3", someHandler, c...)
<-m.Run(9000)
```

#### With REST interface:

```
m := srest.New(nil)
// FriendController implements srest.RESTfuler and generates endpoints for:
// GET     /v1/api/friends
// GET     /v1/api/friends/:id
// POST    /v1/api/friends
// PUT     /v1/api/friends/:id
// DELETE  /v1/api/friends/:id
m.Use("/v1/api/friends", &FriendController{}, Mid1, Mid2, Mid3)
<-m.Run(9000)
```

#### With html templates:

Load Go html templates.
```
// Load templates
err := srest.LoadViews("mytemplatesdir", srest.DefaultFuncMap)
// ...check errors

m := srest.New(nil)
m.Get("/", http.HandlerFunc(homeHandler))
<-m.Run(9000)
```

`mytemplatesdir` files:

```
├── mytemplatesdir
│   └── home.html
```

Using the `home.html` template.

```
func homeHandler(w http.ResponseWriter, r *http.Request) {

    // Call Render func passing variables inside v.
    v := map[string]interface{}{
        "somevar": "A",
    }
    err := srest.Render(w, "home.html", v)
    // ...check errors
}
```

### Payload validation:

```
func Endpoint(w http.ResponseWriter, r *http.Request) {
    r.ParseForm()

    var p Params
    // Bind binds url.Values to struct using gorilla schema
    err := srest.Bind(r.Form, &p)
    // ...check errors
}
```

```
type Modeler interface {
	IsValid() error
}
```

```
// Params implements srest.Modeler interface.
type Params struct{
    Name        string `schema:"name"`
    LastName    string `schema:"last_name"`
}

// IsValid implements srest.Modeler interface.
func(m *Params) IsValid() error {
    if m.Name == "" {
        return errors.New("model: name is required")
    }
    return nil
}
```

Take a look at the working example with all features on examples dir.

### NOTES:

* Prevents errors with endpoint declaration order.

```
m := srest.New(nil)
m.Get("/hello", helloHandler)
m.Get("/hello/:id/friends", helloHandler) // Won't overwrite next endpoint.
m.Get("/hello/:id", helloHandler)
```

* Duplicated endpoints would panic at init time.

```
m := srest.New(nil)
m.Get("/hello", helloHandler)
m.Get("/hello", helloHandler) // Would panic.
m.Get("/hello/:a/name", helloHandler)
m.Get("/hello/:b/name", helloHandler) // Would panic.
<-m.Run(9000)
```

### License:

The MIT License (MIT)

Copyright (c) 2016 Angel del Castillo

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
