package srest

import (
	"errors"
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Model struct satisfies Modeler interface
type Model struct {
	Name string `schema:"name"`
}

// IsValid modeler interface
func (m *Model) IsValid() error {
	return nil
}

func TestBind(t *testing.T) {
	v := url.Values{}
	v.Add("name", "x")
	var x Model
	err := Bind(v, &x)
	assert.Nil(t, err)
}

func TestBindFail(t *testing.T) {
	v := url.Values{}
	v.Add("name", "x")
	var x struct {
		Name string `schema:"name"`
	}
	err := Bind(v, &x)
	assert.NotNil(t, err)
}

// Modelfail struct
type Modelfail struct {
	Name string `schema:"name"`
}

// IsValid modeler interface
func (m *Modelfail) IsValid() error {
	return errors.New("this must fail")
}

func TestModelerFail(t *testing.T) {
	v := url.Values{}
	v.Add("name", "x")
	var x Modelfail
	err := Bind(v, &x)
	assert.EqualValues(t, "this must fail", fmt.Sprintf("%s", err))
}

func TestBindDecoderFail(t *testing.T) {
	v := url.Values{}
	var x Modelfail
	err := Bind(v, x)
	assert.EqualValues(t, "schema: interface must be a pointer to struct", fmt.Sprintf("%s", err))
}
