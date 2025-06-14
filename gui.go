//go:build windows && gui

package main

import (
	"golang.org/x/sys/windows"
	"os"
	"syscall"
)

func init() {
	var s Specs
	if err := s.Collect(); err != nil {
		errBox(err)
		os.Exit(1)
	}
	if err := s.CollectProductKey(); err != nil {
		errBox(err)
		os.Exit(1)
	}

	f, err := s.WriteHTML()
	if err != nil {
		errBox(err)
		os.Exit(1)
	}
	if err := s.OpenHTML(f); err != nil {
		errBox(err)
		os.Exit(1)
	}
}

func errBox(err error) {
	title, _ := syscall.UTF16PtrFromString(err.Error())
	message, _ := syscall.UTF16PtrFromString("Winspecter")

	_, _ = windows.MessageBox(0, title, message, windows.MB_OK|windows.MB_ICONERROR)
}
