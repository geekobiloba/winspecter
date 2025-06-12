//go:build windows && gui

package main

import (
	"golang.org/x/sys/windows"
	"syscall"
)

func main() {
	var s Specs
	if err := s.Collect(); err != nil {
		errorBox(err)
		return
	}
	if err := s.CollectProductKey(); err != nil {
		errorBox(err)
		return
	}
	if err := s.HTMLOpen(); err != nil {
		errorBox(err)
		return
	}
}

func errorBox(err error) {
	title, _ := syscall.UTF16PtrFromString(err.Error())
	message, _ := syscall.UTF16PtrFromString("Winspector Error")

	_, _ = windows.MessageBox(0, title, message, windows.MB_OK|windows.MB_ICONERROR)
}
