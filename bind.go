package srest

import (
	"errors"
	"net/url"

	"github.com/gorilla/schema"
)

// Modeler interface
type Modeler interface {
	IsValid() error
}

var (
	// schDecoder default gorilla schema decoder.
	schDecoder = schema.NewDecoder()

	// ErrImplementsModeler error returned when modeler interface is not implemented.
	ErrImplementsModeler = errors.New("srest: modeler interface not found")
)

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
