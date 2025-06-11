//go:build windows && gui

package main

import (
	"log"
)

func main() {
		var s Specs
		if err := s.Collect(); err != nil {
			log.Fatal(err)
		}

		if err := s.CollectProductKey(); err != nil {
			log.Fatal(err)
		}

		// Write and open HTML file with Windows product key
		// when no argument is given.
		if err := s.CollectProductKey(); err != nil {
			log.Fatal(err)
		}
		if err := s.HTMLOpen(); err != nil {
			log.Fatal(err)
		}
}

