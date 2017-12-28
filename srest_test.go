// Package srest contains tools for sites and web services creation.
/*	Copyright 2016 The SREST Authors. All rights reserved.
	Use of this source code is governed by a BSD-style
	license that can be found in the LICENSE file.
*/
package srest

import (
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
	//	log.Printf("goroutines [%v]", grs)

	if grs > 20 {
		if err := pprof.Lookup("goroutine").WriteTo(os.Stdout, 2); err != nil {
			panic(err)
		}
		panic("exceeded goroutines limit")
	}

	os.Exit(v)
}

// doTempViews creates the html templates required for all tests.
func doTempViews() error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	tmpDir := dir + "/_views"
	log.Printf("TmpViews : tmp dir [%s]", tmpDir)

	if err := os.Remove(tmpDir); err != nil {
		return err
	}

	if err := os.MkdirAll(tmpDir, 0777); err != nil {
		return err
	}

	if err := mkFile(tmpDir+"/index.html", `{{hello world}}`); err != nil {
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
