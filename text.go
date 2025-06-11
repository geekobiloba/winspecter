//go:build windows && cli

package main

import (
	"fmt"
)

func (s *Specs) TextPretty(delim ...string) (out string) {
	tbl := s.Table(s, true, 0)

	const w = 22 // first column wdith

	d := ": " // default column delimiter
	if len(delim) > 0 {
		d = delim[0]
	}

	for _, col := range tbl {
		switch {
		case len(col[1]) > 0:
			out += fmt.Sprintf("%-*s%s%s\n", w, col[0], d, col[1])
		default:
			out += col[0] + d + "\n"
		}
	}

	return
}

func (s *Specs) TextFlat(delim ...string) (out string) {
	tbl := s.Table(s, false, 0)

	const w = 28 // first column wdith

	d := ": " // default column delimiter
	if len(delim) > 0 {
		d = delim[0]
	}

	for _, col := range tbl {
		out += fmt.Sprintf("%-*s%s%s\n", w, col[0], d, col[1])
	}

	return
}

func (s *Specs) TextVCSV(delimQuote ...string) (out string) {
	tbl := s.Table(s, false, 0)

	d, q := ",", `"` // default column delimiter and quote string
	if len(delimQuote) > 0 {
		d = delimQuote[0]
	}
	if len(delimQuote) > 1 {
		q = delimQuote[1]
	}

	for _, col := range tbl {
		out += q + col[0] + q + d + q + col[1] + q + "\n"
	}

	return
}

func (s *Specs) TextCSV(delimQuote ...string) (out string) {
	tbl := s.Table(s, false, 0)

	d, q := ",", `"` // default column delimiter and quote string
	if len(delimQuote) > 0 {
		d = delimQuote[0]
	}
	if len(delimQuote) > 1 {
		q = delimQuote[1]
	}

	l := len(tbl)
	for i := 0; i <= 1; i++ {
		for j := range l {
			out += q + tbl[j][i] + q

			// append delimiter unless the last column
			if j < l-1 {
				out += d
			}
			// append newline to the last column
			if j == l-1 {
				out += "\n"
			}
		}
	}

	return
}

