//go:build windows

package main

import (
	"os/user"
	"regexp"
	"strings"

	"github.com/StackExchange/wmi"
	"golang.org/x/sys/windows/registry"
)

// Specs
// Field order matters here.
// Embedded type names doesn't get processed as parent when being marshaled.
// These tags override it.
type Specs struct {
	CurrentUser `json:"CurrentUser" yaml:"currentuser" toml:"CurrentUser"`
	Windows     `json:"Windows"     yaml:"windows"     toml:"Windows"`
	System      `json:"System"      yaml:"system"      toml:"System"`
	Baseboard   `json:"Baseboard"   yaml:"baseboard"   toml:"Baseboard"`
	BIOS        `json:"BIOS"        yaml:"bios"        toml:"BIOS"`
	CPUs        `json:"CPUs"        yaml:"cpus"        toml:"CPUs"`
	GPUs        `json:"GPUs"        yaml:"gpus"        toml:"GPUs"`
	Memory      `json:"Memory"      yaml:"memory"      toml:"Memory"`
	Disks       `json:"Disks"       yaml:"disks"       toml:"Disks"`
	NetAdapters `json:"NetAdapters" yaml:"netadapters" toml:"NetAdapters"`
}

// Windows
// Following the designation of System > About menu.
type Windows struct {
	CSName             string `json:"DeviceName" yaml:"devicename" toml:"DeviceName"`
	Caption            string `json:"Edition"    yaml:"edition"    toml:"Edition"`
	Version            string
	BuildNumber        string
	SerialNumber       string
	InstallDate        WinInstallDate
	RegisteredUser     string
	OriginalProductKey string
}

type WinInstallDate string

type CurrentUser struct {
	Username string
	Fullname string
	SID      string
}

type CPUs []CPU

type CPU struct {
	Name              string //`json:"Model"       yaml:"model"       toml:"Model"`
	SocketDesignation string `json:"SocketType"  yaml:"sockettype"  toml:"SocketType"`
	NumberOfCores     uint64 `json:"TotalCore"   yaml:"totalcore"   toml:"TotalCore"`
	ThreadCount       uint64 `json:"TotalThread" yaml:"totalthread" toml:"TotalThread"`
	MaxClockSpeed     CPUMaxClockSpeed
	L2CacheSize
	L3CacheSize
}

type CPUMaxClockSpeed uint64
type L2CacheSize uint64
type L3CacheSize uint64

type Memory struct {
	TotalSize DIMMCapacity
	TotalSlot uint64
	DIMMs
}

type DIMMs []DIMM

type DIMM struct {
	DeviceLocator    string
	BankLabel        string
	SMBIOSMemoryType DIMMType `json:"Type" yaml:"type" toml:"Type"`
	Speed            DIMMSpeed
	Capacity         DIMMCapacity
	Manufacturer     string
	PartNumber       string
	SerialNumber     string

	// Not needed for now
	//TypeDetail       DIMMTypeDetail `json:"TypeDetail" yaml:"typedetail" toml:"TypeDetail"`
}

type DIMMType uint64
type DIMMSpeed uint64
type DIMMCapacity uint64

//type DIMMTypeDetail uint64 // not needed for now

type Disks []Disk

type Disk struct {
	//Manufacturer string // not informative
	Model        string
	Size         DiskSize
	SerialNumber string
	Status       string
}

type DiskSize uint64

type GPUs []GPU

type GPU struct {
	Name                 string
	AdapterCompatibility string `json:"Vendor"  yaml:"vendor"  toml:"Vendor"`
	AdapterDACType       string `json:"Type" yaml:"type" toml:"Type"`
}

// BBS
// A helper type
type BBS struct {
	BIOS
	Baseboard
	System
}

type BIOS struct {
	Vendor      string
	Version     string
	ReleaseDate string
}

type Baseboard struct {
	Manufacturer string
	Product      string
	Version      string
}

type System struct {
	Manufacturer string
	Family       string
	Version      string
	ProductName  string
	SKU          string
}

type NetAdapters []NetAdapter

type NetAdapter struct {
	Name         string
	MACAddress   string
	Manufacturer string
}

////////////////////////////////////////////////////////////////////////////////
// Specs
////////////////////////////////////////////////////////////////////////////////

func (s *Specs) Collect() (err error) {

	// Collect current Windows info
	if err = s.Windows.collect(); err != nil {
		return err
	}

	// Collect current user info
	if err = s.CurrentUser.collect(); err != nil {
		return err
	}

	// Collect CPU info
	if err = s.CPUs.collect(); err != nil {
		return err
	}

	// Collect memory info
	if err = s.Memory.collect(); err != nil {
		return err
	}

	// Collect disks info
	if err = s.Disks.collect(); err != nil {
		return err
	}

	// Collect GPUs info
	err = s.GPUs.collect()
	if err != nil {
		return err
	}

	// Collect network adapters info
	if err = s.NetAdapters.collect(); err != nil {
		return err
	}

	// Collect BIOS info
	var bbs BBS
	if err = bbs.collect(); err != nil {
		return err
	}

	s.BIOS, s.Baseboard, s.System = bbs.BIOS, bbs.Baseboard, bbs.System

	return nil
}

func (w *Windows) CollectProductKey() error {
	var k []struct {
		OA3xOriginalProductKey string
	}

	err := wmi.Query(
		"SELECT OA3xOriginalProductKey FROM SoftwareLicensingService",
		&k)
	if err != nil {
		return err
	}

	if k[0].OA3xOriginalProductKey != "" {
		w.OriginalProductKey = k[0].OA3xOriginalProductKey
	} else {
		w.OriginalProductKey = "N/A"
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// Registry Reader
////////////////////////////////////////////////////////////////////////////////

type RegistryReader struct {
	Key registry.Key
}

func NewRegistryReader(path string) (*RegistryReader, error) {
	key, err := registry.OpenKey(
		registry.LOCAL_MACHINE,
		path,
		registry.QUERY_VALUE)
	if err != nil {
		return nil, err
	}

	return &RegistryReader{Key: key}, nil
}

func (r *RegistryReader) GetStringValue(name string) (string, error) {
	val, _, err := r.Key.GetStringValue(name)
	if err != nil {
		return "", err
	}
	return val, nil
}

////////////////////////////////////////////////////////////////////////////////
// CPU
////////////////////////////////////////////////////////////////////////////////

func (c *CPUs) collect() error {
	err := wmi.Query(
		"SELECT Name, SocketDesignation, NumberOfCores, ThreadCount,"+
			"L2CacheSize, L3CacheSize, MaxClockSpeed "+
			"FROM Win32_Processor",
		c)
	if err != nil {
		return err
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// GPU
////////////////////////////////////////////////////////////////////////////////

func (g *GPUs) collect() error {
	err := wmi.Query(
		"SELECT Name, AdapterCompatibility, AdapterDACType "+
			"FROM Win32_VideoController",
		g)
	if err != nil {
		return err
	}

	// Handle empty string values
	for _, s := range *g {
		if s.AdapterCompatibility == "" {
			s.AdapterCompatibility = "N/A"
		}
		if s.AdapterDACType == "" {
			s.AdapterDACType = "N/A"
		}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// Memory
////////////////////////////////////////////////////////////////////////////////

func (m *Memory) collect() error {
	err := wmi.Query(
		"SELECT DeviceLocator, BankLabel, SMBIOSMemoryType, Speed, Capacity, "+
			//"TypeDetail, Manufacturer, PartNumber, SerialNumber " + // not needed for now
			"Manufacturer, PartNumber, SerialNumber "+
			"FROM Win32_PhysicalMemory",
		&m.DIMMs)
	if err != nil {
		return err
	}

	for i := range m.DIMMs {
		m.TotalSize += m.DIMMs[i].Capacity
		m.TotalSlot++

		m.DIMMs[i].PartNumber = strings.TrimSpace(m.DIMMs[i].PartNumber)
	}

	// Handle empty string values
	for _, s := range m.DIMMs {
		if s.Manufacturer == "" {
			s.Manufacturer = "N/A"
		}
		if s.PartNumber == "" {
			s.PartNumber = "N/A"
		}
		if s.SerialNumber == "" {
			s.SerialNumber = "N/A"
		}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// Disks
////////////////////////////////////////////////////////////////////////////////

func (d *Disks) collect() error {
	err := wmi.Query(
		"SELECT Model, Size, SerialNumber, Status FROM Win32_DiskDrive", d)
	if err != nil {
		return err
	}

	// Handle empty string values
	for _, s := range *d {
		if s.Model == "" {
			s.Model = "N/A"
		}
		if s.SerialNumber == "" {
			s.SerialNumber = "N/A"
		}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// Network Adapters
////////////////////////////////////////////////////////////////////////////////

func (n *NetAdapters) collect() error {
	err := wmi.Query(
		"SELECT Name, MACAddress, Manufacturer FROM Win32_NetworkAdapter "+
			"WHERE Manufacturer <> 'Microsoft'",
		n)
	if err != nil {
		return err
	}

	virtAdapters := regexp.MustCompile(`(?i)(Windows|OpenVPN|WireGuard|Oracle|Fortinet)`)
	realAdapters := (*n)[:0] // same capacity as *n, no reallocation

	for _, v := range *n {
		// Add adapter to realAdapters if not matches virtAdapter
		if !virtAdapters.MatchString(v.Manufacturer) {
			realAdapters = append(realAdapters, v)
		}
	}
	*n = realAdapters

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// Current User
////////////////////////////////////////////////////////////////////////////////

func (u *CurrentUser) collect() error {
	v, err := user.Current()
	if err != nil {
		return err
	}

	u.Username, u.Fullname, u.SID = v.Username, v.Name, v.Uid

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// Windows info
////////////////////////////////////////////////////////////////////////////////

func (w *Windows) collect() error {

	// Collect main Windows info
	var v []struct {
		CSName         string
		Caption        string
		BuildNumber    string
		SerialNumber   string
		InstallDate    WinInstallDate
		RegisteredUser string
	}

	err := wmi.Query(
		"SELECT Caption, BuildNumber, SerialNumber, CSName, InstallDate,"+
			"RegisteredUser "+
			"FROM Win32_OperatingSystem",
		&v)
	if err != nil {
		return err
	}

	*w = Windows{
		CSName:             v[0].CSName,
		Caption:            v[0].Caption,
		BuildNumber:        v[0].BuildNumber,
		SerialNumber:       v[0].SerialNumber,
		InstallDate:        v[0].InstallDate,
		RegisteredUser:     v[0].RegisteredUser,
		OriginalProductKey: "***********", // default; handled by its own method
	}

	// Collect Windows feature update version, e.g. 24H2
	reg, err := NewRegistryReader(`SOFTWARE\Microsoft\Windows NT\CurrentVersion`)
	if err != nil {
		return err
	}
	defer func(Key registry.Key) {
		err := Key.Close()
		if err != nil {
			return
		}
	}(reg.Key)

	w.Version, err = reg.GetStringValue("DisplayVersion")
	if err != nil {
		return err
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// BIOS, Baseboard, System (BBS)
////////////////////////////////////////////////////////////////////////////////

func (b *BBS) collect() error {
	reg, err := NewRegistryReader(`HARDWARE\Description\System\BIOS`)

	if err != nil {
		return err
	}
	defer func(Key registry.Key) {
		err := Key.Close()
		if err != nil {
			return
		}
	}(reg.Key)

	b.BIOS.Vendor, err = reg.GetStringValue("BIOSVendor")
	if err != nil {
		return err
	}
	if b.BIOS.Vendor == "" {
		b.BIOS.Vendor = "N/A"
	}

	b.BIOS.Version, err = reg.GetStringValue("BIOSVersion")
	if err != nil {
		return err
	}
	if b.BIOS.Version == "" {
		b.BIOS.Version = "N/A"
	}

	b.BIOS.ReleaseDate, err = reg.GetStringValue("BIOSReleaseDate")
	if err != nil {
		return err
	}
	if b.BIOS.ReleaseDate == "" {
		b.BIOS.ReleaseDate = "N/A"
	}

	b.Baseboard.Manufacturer, err = reg.GetStringValue("BaseBoardManufacturer")
	if err != nil {
		return err
	}
	if b.Baseboard.Manufacturer == "" {
		b.Baseboard.Manufacturer = "N/A"
	}

	b.Baseboard.Product, err = reg.GetStringValue("BaseBoardProduct")
	if err != nil {
		return err
	}
	if b.Baseboard.Product == "" {
		b.Baseboard.Product = "N/A"
	}

	b.Baseboard.Version, err = reg.GetStringValue("BaseBoardVersion")
	if err != nil {
		return err
	}
	if b.Baseboard.Version == "" {
		b.Baseboard.Version = "N/A"
	}

	b.System.Manufacturer, err = reg.GetStringValue("SystemManufacturer")
	if err != nil {
		return err
	}
	if b.System.Manufacturer == "" {
		b.System.Manufacturer = "N/A"
	}

	b.System.Family, err = reg.GetStringValue("SystemFamily")
	if err != nil {
		return err
	}
	if b.System.Family == "" {
		b.System.Family = "N/A"
	}

	b.System.Version, err = reg.GetStringValue("SystemVersion")
	if err != nil {
		return err
	}
	if b.System.Version == "" {
		b.System.Version = "N/A"
	}

	b.System.ProductName, err = reg.GetStringValue("SystemProductName")
	if err != nil {
		return err
	}
	if b.System.ProductName == "" {
		b.System.ProductName = "N/A"
	}

	b.System.SKU, err = reg.GetStringValue("SystemSKU")
	if err != nil {
		return err
	}
	if b.System.SKU == "" {
		b.System.SKU = "N/A"
	}

	return nil
}
