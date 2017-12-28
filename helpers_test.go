package srest

import (
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSON(t *testing.T) {
	w := httptest.NewRecorder()
	err := JSON(w, `this is string`)
	assert.Nil(t, err)

	s := w.Body.String()

	expected := []byte(`"this is string"`)
	actual := s[:len(s)-1]
	assert.EqualValues(t, expected, actual)
}
