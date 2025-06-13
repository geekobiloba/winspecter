//go:build windows && gui

package main

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"html/template"
	"os"
	"os/exec"
	"regexp"
	"time"
)

type tag struct {
	h1
	h2
	h3
	section
	table
}
type h1 struct{ begin, end string }
type h2 struct{ begin, end string }
type h3 struct{ begin, end string }
type section struct{ begin, end string }
type table struct {
	begin, end string
	tbody
	tr
	td
}
type tbody struct{ begin, end string }
type tr struct{ begin, end string }
type td struct{ begin, end string }

var t = tag{
	h1: h1{
		begin: `  <h1>`,
		end:   `</h1>`,
	},
	h2: h2{
		begin: `  <h2>`,
		end:   `</h2>`,
	},
	h3: h3{
		begin: `  <h3>`,
		end:   `</h3>`,
	},
	section: section{
		begin: `<section>`,
		end:   `</section>`,
	},
	table: table{
		begin: `  <table>`,
		end:   `  </table>`,
		tbody: tbody{
			begin: `    <tbody>`,
			end:   `    </tbody>`,
		},
		tr: tr{
			begin: `      <tr>`,
			end:   `</tr>`,
		},
		td: td{
			begin: `<td>`,
			end:   `</td>`,
		},
	},
}

func (s *Specs) genHTMLBody() (out string) {
	table := s.Table(s, true, 0)
	leads := regexp.MustCompile(`^\s+`)
	level := []*regexp.Regexp{
		regexp.MustCompile(`^\S`),      // level 0 is unindented,
		regexp.MustCompile(`^\s{2}\S`), // level 1 is 2-space indented,
		regexp.MustCompile(`^\s{4}\S`), // level 2 is 4-space indented,
		regexp.MustCompile(`^\s{6}\S`), // level 3 is 6-space indented.
	}

	var tableBeginAllowed bool

	for i, col := range table {
		switch {
		case level[0].MatchString(col[0]):
			if i == 0 {
				out += t.section.begin + "\n"
				out += t.h1.begin + col[0] + t.h1.end + "\n"
			} else {
				out += t.table.end + "\n"
				out += t.section.end + "\n"
				out += t.section.begin + "\n"
				out += t.h1.begin + col[0] + t.h1.end + "\n"
			}

			tableBeginAllowed = true

		default:
			j := 1
			if level[2].MatchString(col[0]) {
				j = 2
			}

			str := leads.ReplaceAllString(col[0], "")

			if col[1] == "" {
				if !tableBeginAllowed {
					out += t.table.tbody.end + "\n"
					out += t.table.end + "\n"

					tableBeginAllowed = true
				}

				out += map[int]string{
					1: t.h2.begin + str + t.h2.end + "\n",
					2: t.h3.begin + str + t.h3.end + "\n",
				}[j]

			} else {
				if tableBeginAllowed {
					out += t.table.begin + "\n"
					out += t.table.tbody.begin + "\n"

					tableBeginAllowed = false
				}

				out += t.table.tr.begin
				out += t.table.td.begin + str + t.table.td.end
				out += t.table.td.begin + col[1] + t.table.td.end
				out += t.table.tr.end + "\n"
			}
		}
	}

	out += t.table.tbody.end + "\n"
	out += t.table.end + "\n"
	out += t.section.end

	return out
}

//go:embed assets/html.tmpl
var htmlTmpl string

//go:embed assets/style.css
var htmlCSS string

//go:embed assets/script.js
var htmlJS string

//go:embed assets/favicon.ico
var favicon []byte

var htmlTimestamp = time.Now()

func (s *Specs) genHTMLFull() (string, error) {
	htmlData := map[string]any{
		"body":      template.HTML(s.genHTMLBody()),
		"css":       template.CSS(htmlCSS),
		"js":        template.JS(htmlJS),
		"icon":      base64.StdEncoding.EncodeToString(favicon),
		"version":   Version,
		"timestamp": htmlTimestamp.Format("Mon, 02 Jan 2006 15:04:05 UTC-0700"),
	}

	t := template.Must(template.New("htmlpage").Parse(htmlTmpl))

	// Apply template and buffer to buff
	var buff bytes.Buffer
	err := t.Execute(&buff, htmlData)
	if err != nil {
		return "", err
	}

	return buff.String(), nil
}

func (s *Specs) WriteHTML() (filename string, err error) {
	re := regexp.MustCompile(`([^\\]+)\\([^\\]+)`)

	userAtHost := re.ReplaceAllString(s.CurrentUser.Username, "$2@$1")
	timestamp := htmlTimestamp.Format("20060102T150405-0700")

	filename = userAtHost + "_" + timestamp + ".html"

	t, err := s.genHTMLFull()
	if err != nil {
		return "", err
	}

	// Write the HTML file.
	// Permission 600 will make it readonly.
	if err = os.WriteFile(filename, []byte(t), 700); err != nil {
		return "", err
	}

	return filename, nil
}

func (s *Specs) OpenHTML(filename string) error {
	// Check if filename exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return err
	}
	// Open the HTML file with the default browser.
	cmd := exec.Command("rundll32", "url.dll,FileProtocolHandler", filename)
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
