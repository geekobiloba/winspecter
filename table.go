//go:build windows

package main

import (
	"fmt"
	"reflect"
)

// Convert the specs struct as a 2-column table for later processing.
// If pretty is true, indentation is added to the first column
// for easier pretty printing.
// Otherwise, the struct is flattened to make a transposed CSV.
// Normal CSV creation, its column delimiter and quote character
// should be handled by the consumer.

func (s *Specs) Table(data any, pretty bool, a int, label ...string) (z [][]string) {
	t := reflect.TypeOf(data)
	v := reflect.ValueOf(data)

	var l string
	if len(label) > 0 {
		l = label[0]
	}

	b, m := a, l // recursion keepers

	const w = 2  // indentation width per level.
	const x = "" // indentation helper string.

	if t.Kind() == reflect.Ptr {
		t, v = t.Elem(), v.Elem()
	}

	if t.Kind() != reflect.Struct {
		return
	}

	for i := range t.NumField() {
		key, val := t.Field(i), v.Field(i)

		switch val.Kind() {
		case reflect.Struct:
			switch {
			case pretty:
				a = b
				z = append(z, []string{fmt.Sprintf("%-*s%s", a, x, key.Name), ""})

				a += w // a = b+w works, too; probably needs further analysis.
			default:
				a, l = 0, key.Name+" "
			}

			// mind the elipsis ...
			z = append(z, s.Table(val.Interface(), pretty, a, l)...)

		case reflect.Slice:
			if pretty {
				l = key.Name
				a = b
				z = append(z, []string{fmt.Sprintf("%-*s%s", a, x, l), ""})
			}

			for j := range val.Len() {
				l = fmt.Sprintf("%s%d", val.Index(j).Type().Name(), j)

				switch {
				case pretty:
					a = b + w
					z = append(z, []string{fmt.Sprintf("%-*s%s", a, x, l), ""})

					a, l = b+2*w, ""
				default:
					a, l = b, l+" "
				}

				// mind the elipsis ...
				z = append(z, s.Table(val.Index(j).Interface(), pretty, a, l)...)
			}

		default:
			r := val.Interface()

			switch {
			case pretty:
				l = s.MapKey(key.Name)
			default:
				l = m + s.MapKey(key.Name)
			}

			z = append(z, []string{
				fmt.Sprintf("%-*s%s", a, x, l),
				fmt.Sprintf("%v", r),
			})
		}
	}

	return
}

// MapKey Translate key names to be more intuitive.
func (*Specs) MapKey(f string) string {
	// This map is unsafe.
	// Two children of different parents may have the same field name.
	m := map[string]string{
		"CSName":               "DeviceName",
		"Caption":              "Edition",
		"SocketDesignation":    "SocketType",
		"NumberOfCores":        "TotalCore",
		"ThreadCount":          "TotalThread",
		"SMBIOSMemoryType":     "Type",
		"AdapterCompatibility": "Vendor",
		"AdapterDACType":       "Type",
	}

	if g, ok := m[f]; ok {
		return g
	}
	return f
}
