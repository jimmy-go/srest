// Package srest contains test for bug fixes.
/*	Copyright 2016 The SREST Authors. All rights reserved.
	Use of this source code is governed by a BSD-style
	license that can be found in the LICENSE file.
*/
package srest

import (
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBugRaceRender(t *testing.T) {
	dir, err := getTempDir()
	assert.Nil(t, err)

	err = LoadViews(dir+"/a", DefaultFuncMap)
	assert.Nil(t, err)

	racerender(t, 1000)
}

func TestBugRaceRenderDebug(t *testing.T) {
	dir, err := getTempDir()
	assert.Nil(t, err)

	err = LoadViews(dir+"/a", DefaultFuncMap)
	assert.Nil(t, err)

	Debug(true)
	racerender(t, 1000)
}

func racerender(t *testing.T, l int) {
	var wg sync.WaitGroup
	wg.Add(l)
	for i := 0; i < l; i++ {
		go func() {
			defer wg.Done()
			w := httptest.NewRecorder()
			err := Render(w, "index.html", map[string]interface{}{"x": 1})
			assert.Nil(t, err)

			actual := w.Body.String()
			expected := "I am lowercase-eqs:true"
			assert.EqualValues(t, expected, actual)
		}()
	}
	wg.Wait()
}

// TestBugAllViewsLoaded demonstrates all views are loaded.
func TestBugAllViewsLoaded(t *testing.T) {
	dir, err := getTempDir()
	assert.Nil(t, err)

	err = LoadViews(dir+"/a", DefaultFuncMap)
	assert.Nil(t, err)

	table := []struct {
		Purpose  string
		Name     string
		ExpBody  string
		ExpError error
	}{
		{
			"1. OK",
			"all/all.html",
			`before_index::I am lowercase-eqs:true::after_index.before_menu::menu::after_menu`,
			nil,
		},
		{
			"2. OK",
			"index.html",
			`I am lowercase-eqs:true`,
			nil,
		},
		{
			"3. OK",
			"menu.html",
			`menu`,
			nil,
		},
	}
	for i := range table {
		x := table[i]

		w := httptest.NewRecorder()
		err := Render(w, x.Name, map[string]interface{}{"x": 1})
		assert.EqualValues(t, x.ExpError, err, x.Purpose)

		actual := w.Body.String()
		assert.EqualValues(t, x.ExpBody, actual, x.Purpose)
	}
}
