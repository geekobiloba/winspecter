//go:build windows && cli

package main

import (
	"flag"
	"fmt"
	"log"
)

func main() {

	//****************************************************************************
	// Usage Text
	//****************************************************************************

	usageText := map[string][]string{
		"header": {
			"Winspecter - Win Specs Reporter",
			"Options:",
		},
		"footer": {
			"Notes:\n",
			"  Use the launcher to generate HTML in current directory.",
		},
	}

	flag.Usage = func() {
		for _, line := range usageText["header"] {
			_, _ = fmt.Fprintf(flag.CommandLine.Output(), "%s\n\n", line)
		}

		flag.PrintDefaults()

		_, _ = fmt.Fprintln(flag.CommandLine.Output(), "")
		for _, line := range usageText["footer"] {
			_, _ = fmt.Fprintln(flag.CommandLine.Output(), line)
		}
	}

	//****************************************************************************
	// Options
	//****************************************************************************

	actions := map[string]*bool{
		"json":   flag.Bool("json", false, "Print as JSON."),
		"yaml":   flag.Bool("yaml", false, "Print as YAML."),
		"toml":   flag.Bool("toml", false, "Print as TOML."),
		"pretty": flag.Bool("pretty", false, "Pretty print (YAML-like)."),
		"print":  flag.Bool("print", false, "Alias for pretty print."),
		"flat":   flag.Bool("flat", false, "Print as flat list."),
		"csv":    flag.Bool("csv", false, "Print as CSV."),

		"vcsv": flag.Bool("vcsv", false,
			"Print as vertical/transposed CSV "+
				"(headers in rows, instead of single row)."),

		"version": flag.Bool("version", false, "Print version."),
	}

	// CSV and VCSV flags
	quote := flag.String("quote", `"`, "Quote string for CSV.")
	delim := flag.String("delim", `,`, "CSV and VCSV column delimiter.")

	// All format flags
	withKey := flag.Bool("key", false, "Include Windows product key.")

	//****************************************************************************
	// Parse Args
	//****************************************************************************

	flag.Parse()

	var selectedAction string
	for action, selected := range actions {
		if *selected {
			selectedAction = action
			break
		}
	}

	switch {

	// Handle absent and "naked" or invalid args
	case selectedAction == "", len(flag.Args()) > 0:
		flag.Usage()
		return

	// Print version
	case selectedAction == "version":
		fmt.Printf("Winspecter v%s\n", Version)
		return
	}

	var s Specs
	if err := s.Collect(); err != nil {
		log.Fatal(err)
	}

	if *withKey {
		if err := s.CollectProductKey(); err != nil {
			log.Fatal(err)
		}
	}

	switch selectedAction {
	case "json":
		res, err := s.JSON()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(res)

	case "yaml":
		res, err := s.YAML()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(res)

	case "toml":
		res, err := s.TOML()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(res)

	case "pretty", "print":
		fmt.Println(s.TextPretty(": "))

	case "flat":
		fmt.Println(s.TextFlat(": "))

	case "csv":
		fmt.Println(s.TextCSV(*delim, *quote))

	case "vcsv":
		fmt.Println(s.TextVCSV(*delim, *quote))

	// Just in case
	default:
		flag.Usage()
	}
}
