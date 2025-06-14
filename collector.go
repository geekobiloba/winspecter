//go:build windows

package main

import (
	"context"
	"os/user"
	"strings"
	"time"

	"github.com/yusufpapurcu/wmi"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sys/windows/registry"
	//"github.com/StackExchange/wmi"
	//"github.com/microsoft/wmi/go/wmi"
)

// Specs
// Field order matters here.
// Embedded type names don't get processed as parent when being marshaled.
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

const wmiTimeout = 5 * time.Second

////////////////////////////////////////////////////////////////////////////////
// Public Methods
////////////////////////////////////////////////////////////////////////////////

func (s *Specs) Collect() (err error) {
	bbs := &BBS{}
	ctx := context.Background()
	g, _ := errgroup.WithContext(ctx)

	g.Go(func() error {
		return s.Windows.collect()
	})
	g.Go(func() error {
		return s.CurrentUser.collect()
	})
	g.Go(func() error {
		return s.CPUs.collect()
	})
	g.Go(func() error {
		return s.GPUs.collect()
	})
	g.Go(func() error {
		return s.Memory.collect()
	})
	g.Go(func() error {
		return s.Disks.collect()
	})
	g.Go(func() error {
		return s.NetAdapters.collect()
	})
	g.Go(func() error {
		return bbs.collect()
	})

	if err := g.Wait(); err != nil {
		return err
	}

	s.BIOS, s.Baseboard, s.System = bbs.BIOS, bbs.Baseboard, bbs.System

	return nil
}

func (w *Windows) CollectProductKey() error {
	var k []struct {
		OA3xOriginalProductKey string
	}

	ctx, cancel := context.WithTimeout(context.Background(), wmiTimeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- wmi.Query(
			"SELECT OA3xOriginalProductKey FROM SoftwareLicensingService",
			&k)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		if err != nil {
			return err
		}
	}

	if k[0].OA3xOriginalProductKey != "" {
		w.OriginalProductKey = k[0].OA3xOriginalProductKey
	} else {
		w.OriginalProductKey = "N/A"
	}

	return nil
}

//******************************************************************************
// Registry Reader
//******************************************************************************

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
	ctx, cancel := context.WithTimeout(context.Background(), wmiTimeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- wmi.Query(
			"SELECT Name, SocketDesignation, NumberOfCores, ThreadCount, "+
				"L2CacheSize, L3CacheSize, MaxClockSpeed "+
				"FROM Win32_Processor",
			c)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		if err != nil {
			return err
		}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// GPU
////////////////////////////////////////////////////////////////////////////////

func (g *GPUs) collect() error {
	ctx, cancel := context.WithTimeout(context.Background(), wmiTimeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- wmi.Query(
			"SELECT Name, AdapterCompatibility, AdapterDACType "+
				"FROM Win32_VideoController",
			g)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		if err != nil {
			return err
		}
	}

	// Handle empty string
	for i := range *g {
		if (*g)[i].AdapterCompatibility == "" {
			(*g)[i].AdapterCompatibility = "N/A"
		}
		if (*g)[i].AdapterDACType == "" {
			(*g)[i].AdapterDACType = "N/A"
		}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// Memory
////////////////////////////////////////////////////////////////////////////////

func (m *Memory) collect() error {
	ctx, cancel := context.WithTimeout(context.Background(), wmiTimeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- wmi.Query(
			"SELECT DeviceLocator, BankLabel, SMBIOSMemoryType, Speed, Capacity, "+
				//"TypeDetail, Manufacturer, PartNumber, SerialNumber " + // not needed for now
				"Manufacturer, PartNumber, SerialNumber "+
				"FROM Win32_PhysicalMemory",
			&m.DIMMs)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		if err != nil {
			return err
		}
	}

	for i := range m.DIMMs {
		m.TotalSize += m.DIMMs[i].Capacity
		m.TotalSlot++

		m.DIMMs[i].PartNumber = strings.TrimSpace(m.DIMMs[i].PartNumber)
	}

	// Handle empty string
	for i := range m.DIMMs {
		if m.DIMMs[i].Manufacturer == "" {
			m.DIMMs[i].Manufacturer = "N/A"
		}
		if m.DIMMs[i].PartNumber == "" {
			m.DIMMs[i].PartNumber = "N/A"
		}
		if m.DIMMs[i].SerialNumber == "" {
			m.DIMMs[i].SerialNumber = "N/A"
		}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// Disks
////////////////////////////////////////////////////////////////////////////////

func (d *Disks) collect() error {
	ctx, cancel := context.WithTimeout(context.Background(), wmiTimeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- wmi.Query(
			"SELECT Model, Size, SerialNumber, Status FROM Win32_DiskDrive",
			d)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		if err != nil {
			return err
		}
	}

	// Handle empty string
	for i := range *d {
		if (*d)[i].Model == "" {
			(*d)[i].Model = "N/A"
		}
		if (*d)[i].SerialNumber == "" {
			(*d)[i].SerialNumber = "N/A"
		}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////
// Network Adapters
////////////////////////////////////////////////////////////////////////////////

func (n *NetAdapters) collect() error {
	ctx, cancel := context.WithTimeout(context.Background(), wmiTimeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- wmi.Query(
			"SELECT Name, MACAddress, Manufacturer FROM Win32_NetworkAdapter "+
				"WHERE Manufacturer <> 'Microsoft'",
			n)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		if err != nil {
			return err
		}
	}

	realAdapters := (*n)[:0] // same capacity as *n, no reallocation
	virtAdapters := []string{
		"windows",
		"openvpn",
		"wireguard",
		"oracle",
		"fortinet",
	}

outer:
	for _, v := range *n {
		m := strings.ToLower(v.Manufacturer)

		for _, w := range virtAdapters {
			if strings.Contains(m, w) {
				continue outer
			}
		}
		realAdapters = append(realAdapters, v)
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
	// wmi.Query receiver struct must have exact fields as the queried ones,
	// so here's a temporary struct to hold the results.
	var v []struct {
		CSName         string
		Caption        string
		BuildNumber    string
		SerialNumber   string
		InstallDate    WinInstallDate
		RegisteredUser string
	}

	ctx, cancel := context.WithTimeout(context.Background(), wmiTimeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- wmi.Query(
			"SELECT Caption, BuildNumber, SerialNumber, CSName, InstallDate,"+
				"RegisteredUser "+
				"FROM Win32_OperatingSystem",
			&v)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		if err != nil {
			return err
		}
	}

	*w = Windows{
		CSName:             v[0].CSName,
		Caption:            v[0].Caption,
		BuildNumber:        v[0].BuildNumber,
		SerialNumber:       v[0].SerialNumber,
		InstallDate:        v[0].InstallDate,
		RegisteredUser:     v[0].RegisteredUser,
		OriginalProductKey: "***********", // default when is not requested
	}

	// Collect Windows feature update version, e.g., 24H2
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
