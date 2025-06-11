//go:build windows

package main

import (
	"os/user"
	"strings"

	"github.com/StackExchange/wmi"
	"golang.org/x/sys/windows/registry"
)

// Specs Field order matters here,
//
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
}

// Windows Following the designation of System > About menu.
type Windows struct {
	CSName             string `json:"DeviceName" yaml:"devicename" toml:"DeviceName"`
	Caption            string `json:"Edition"    yaml:"edition"    toml:"Edition"`
	Version            string
	BuildNumber        string
	SerialNumber       string
	OriginalProductKey string
	InstallDate        WinInstallDate
	RegisteredUser     string
}

type WinInstallDate string

type CurrentUser struct {
	Username string
	Realname string
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

// BBS A helper type
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

////////////////////////////////////////////////////////////////////////////////
// Specs
////////////////////////////////////////////////////////////////////////////////

func (s *Specs) Collect() (err error) {

	// Collect current Windows info
	if err = s.Windows.Collect(); err != nil {
		return err
	}

	// Collect current user info
	if err = s.CurrentUser.Collect(); err != nil {
		return err
	}

	// Collect CPU info
	if err = s.CPUs.Collect(); err != nil {
		return err
	}

	// Collect memory info
	if err = s.Memory.Collect(); err != nil {
		return err
	}

	// Collect disks info
	if err = s.Disks.Collect(); err != nil {
		return err
	}

	// Collect GPUs info
	err = s.GPUs.Collect()
	if err != nil {
		return err
	}

	// Collect BIOS info
	//bbs := &BBS{} // GCed
	var bbs BBS
	if err = bbs.Collect(); err != nil {
		return err
	}

	s.BIOS, s.Baseboard, s.System = bbs.BIOS, bbs.Baseboard, bbs.System

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

func (c *CPUs) Collect() error {
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

func (g *GPUs) Collect() error {
	err := wmi.Query(
		"SELECT Name, AdapterCompatibility, AdapterDACType "+
			"FROM Win32_VideoController",
		g)
	if err != nil {
		return err
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// Memory
////////////////////////////////////////////////////////////////////////////////

func (m *Memory) Collect() error {
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

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// Disks
////////////////////////////////////////////////////////////////////////////////

func (d *Disks) Collect() error {
	err := wmi.Query(
		"SELECT Model, Size, SerialNumber, Status FROM Win32_DiskDrive", d)
	if err != nil {
		return err
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
// Current User
////////////////////////////////////////////////////////////////////////////////

func (u *CurrentUser) Collect() error {
	v, err := user.Current()
	if err != nil {
		return err
	}

	u.Username, u.Realname, u.SID = v.Username, v.Name, v.Uid

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// Windows info
////////////////////////////////////////////////////////////////////////////////

func (w *Windows) Collect() error {

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

	w.OriginalProductKey = k[0].OA3xOriginalProductKey

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// BIOS, Baseboard, System (BBS)
////////////////////////////////////////////////////////////////////////////////

func (b *BBS) Collect() error {
	reg, err := NewRegistryReader(`HARDWARE\Description\System\BIOS`)

	if err != nil {
		return err
	}
	defer func(Key registry.Key) {
		if err := Key.Close(); err != nil {
			return
		}
	}(reg.Key)

	b.BIOS.Vendor, err = reg.GetStringValue("BIOSVendor")
	if err != nil {
		return err
	}

	b.BIOS.Version, err = reg.GetStringValue("BIOSVersion")
	if err != nil {
		return err
	}

	b.BIOS.ReleaseDate, err = reg.GetStringValue("BIOSReleaseDate")
	if err != nil {
		return err
	}

	b.Baseboard.Manufacturer, err = reg.GetStringValue("BaseBoardManufacturer")
	if err != nil {
		return err
	}

	b.Baseboard.Product, err = reg.GetStringValue("BaseBoardProduct")
	if err != nil {
		return err
	}

	b.Baseboard.Version, err = reg.GetStringValue("BaseBoardVersion")
	if err != nil {
		return err
	}

	b.System.Manufacturer, err = reg.GetStringValue("SystemManufacturer")
	if err != nil {
		return err
	}

	b.System.Family, err = reg.GetStringValue("SystemFamily")
	if err != nil {
		return err
	}

	b.System.Version, err = reg.GetStringValue("SystemVersion")
	if err != nil {
		return err
	}

	b.System.ProductName, err = reg.GetStringValue("SystemProductName")
	if err != nil {
		return err
	}

	b.System.SKU, err = reg.GetStringValue("SystemSKU")
	if err != nil {
		return err
	}

	return nil
}
