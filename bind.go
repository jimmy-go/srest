// Package srest contains tools for REST services and web sites.
/*	Copyright 2016 The SREST Authors. All rights reserved.
	Use of this source code is governed by a BSD-style
	license that can be found in the LICENSE file.
*/
package srest

import (
	"errors"
	"net/url"

	"github.com/gorilla/schema"
)

var (
	// schDecoder default gorilla schema decoder.
	schDecoder = schema.NewDecoder()

	// ErrImplementsModeler error returned when modeler interface is not implemented.
	ErrImplementsModeler = errors.New("srest: modeler interface not found")
)

// Modeler interface
type Modeler interface {
	IsValid() error
}

// Bind implements gorilla schema and runs IsValid method from data.
func Bind(vars url.Values, dst interface{}) error {
	err := schDecoder.Decode(dst, vars)
	if err != nil {
		return err
	}
	// check model is valid
	mo, ok := dst.(Modeler)
	if !ok {
		return ErrImplementsModeler
	}
	return mo.IsValid()
}
