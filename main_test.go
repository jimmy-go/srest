// Package srest contains tools for sites and web services creation.
/*	Copyright 2016 The SREST Authors. All rights reserved.
	Use of this source code is governed by a BSD-style
	license that can be found in the LICENSE file.
*/
package srest

import (
	"errors"
	"io"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"testing"
)

func TestMain(m *testing.M) {
	// Create temporal dir and templates.
	if err := doTempViews(); err != nil {
		panic(err)
	}

	v := m.Run()

	grs := runtime.NumGoroutine()
	log.Printf("goroutines [%v]", grs)

	if grs > 20 {
		if err := pprof.Lookup("goroutine").WriteTo(os.Stdout, 2); err != nil {
			panic(err)
		}
		panic("exceeded goroutines limit")
	}

	os.Exit(v)
}

const (
	tmpDirName = "_tmp_views"
)

// getTempDir returns the temporal dir for templates tests.
func getTempDir() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	s := dir + "/" + tmpDirName
	return s, nil
}

// doTempViews creates the html templates required for all tests.
func doTempViews() error {
	tmpDir, err := getTempDir()
	if err != nil {
		return err
	}
	log.Printf("TmpViews : tmp dir [%s]", tmpDir)

	dirs := []struct {
		Dir string
	}{
		{tmpDir},
		{tmpDir + "/a/all"},
		{tmpDir + "/b"},
		{tmpDir + "/c"},
	}
	for _, x := range dirs {
		if err := mkDir(x.Dir); err != nil {
			return err
		}
	}

	files := []struct {
		File, Content string
	}{
		{
			"a/all/all.html",
			`before_index::{{template "index.html" .}}::after_index.before_menu::{{template "menu.html" .}}::after_menu`,
		},
		{"a/index.html", `{{cap "i am lowercase"}}-eqs:{{eqs 1 "1"}}{{cap ""}}`},
		{"a/menu.html", `menu`},
		{"b/sideb.html", `this is side B`},
		{"c/empty.html", ``},
	}
	for _, x := range files {
		if err := mkFile(tmpDir+"/"+x.File, x.Content); err != nil {
			return err
		}
	}

	return nil
}

func mkDir(path string) error {
	if path == "" {
		return errors.New("empty path")
	}
	if err := os.RemoveAll(path); err != nil {
		return err
	}
	if err := os.MkdirAll(path, 0777); err != nil {
		return err
	}
	return nil
}

func mkFile(path, body string) error {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	if _, err := io.WriteString(f, body); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return nil
}
