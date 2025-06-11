//go:build windows

package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/docker/go-units"
)

////////////////////////////////////////////////////////////////////////////////
// CPU
////////////////////////////////////////////////////////////////////////////////

//func (c CPUMaxClockSpeed) String() string {
//  return fmt.Sprintf("%.1f GHz", float64(c)/1e3)
//}

func (c CPUMaxClockSpeed) String() string {
	return fmt.Sprintf("%.3f", float64(c)/1e3) // 3 decimal digits
}

// WMI returns L2 and L3 cache size in KiB

func (c L2CacheSize) String() string {
	return fmt.Sprintf("%d", int64(c)/units.KiB)
}
func (c L3CacheSize) String() string {
	return fmt.Sprintf("%d", int64(c)/units.KiB)
}

////////////////////////////////////////////////////////////////////////////////
// Memory
////////////////////////////////////////////////////////////////////////////////

// See: https://www.dmtf.org/sites/default/files/standards/documents/DSP0134_3.4.0.pdf
// Table 77
func (d DIMMType) String() string {
	switch d {
	case 18:
		return "DDR"
	case 19:
		return "DDR2"
	case 20:
		return "DDR2 FB-DIMM"
	case 24:
		return "DDR3"
	case 26:
		return "DDR4"
	case 27:
		return "LPDDR"
	case 28:
		return "LPDDR2"
	case 29:
		return "LPDDR3"
	case 30:
		return "LPDDR4"
	case 34:
		return "DDR5"
	case 35:
		return "LPDDR5"
	default:
		return "unknown"
	}
}

//func (d DIMMSpeed) String() string {
//  return fmt.Sprintf("%d MT/s", d)
//}

//func (d DIMMCapacity) String() string {
//  return fmt.Sprintf("%d GiB", d/units.GiB)
//}

func (d DIMMCapacity) String() string {
	return fmt.Sprintf("%d", d/units.GiB)
}

// Not needed for now
/*
// See: https://www.dmtf.org/sites/default/files/standards/documents/DSP0134_3.4.0.pdf
// Table 78
func (d DIMMTypeDetail) String() (s string) {
	// Not needed, as Windows runs only on Synchronous memory
	// Bit 7 --> Synchronous
	//switch d>>7 - 128 {
	//case 1:
	//	s += "Synchronous"
	//case 0:
	//	s += "Asynchronous" // is this the correct term?
	//default:
	//	s += "unknown"
	//}
	//s += ", "

	switch d>>13 {
	case 1:
		s += "Registered (Buffered)"
	case 2:
		s += "Unbuffered (Unregistered)" // sic!
	default:
		s += "unknown"
	}
	//returnfmt.Sprintf("%d", d>>7 - 128 )
	return s
}
//*/

////////////////////////////////////////////////////////////////////////////////
// Disk
////////////////////////////////////////////////////////////////////////////////

//func (s DiskSize) String() string {
//  return fmt.Sprintf("%d GB", s/units.GB)
//}

func (d DiskSize) String() string {
	return fmt.Sprintf("%d", d/units.GB)
}

////////////////////////////////////////////////////////////////////////////////
// Windows
////////////////////////////////////////////////////////////////////////////////

func (d WinInstallDate) String() string {
	datetime := string(d[:14])
	offset := string(d[22:])

	offsetMin, err := strconv.Atoi(offset)
	if err != nil {
		return fmt.Sprintf("%s", err.Error())
	}
	offsetDur := time.Duration(offsetMin) * time.Minute

	h := int(offsetDur.Hours())
	m := int(offsetDur.Minutes()) % 60
	z := fmt.Sprintf("%s%+03d%02d", datetime, h, m)

	t, err := time.Parse("20060102150405-0700", z)
	if err != nil {
		return fmt.Sprintf("%s", err.Error())
	}
	return t.Format(time.RFC3339)
}
