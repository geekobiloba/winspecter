//go:build windows && cli

package main

import (
	"encoding/json"
	"github.com/BurntSushi/toml"
	"github.com/docker/go-units"
	"gopkg.in/yaml.v3"
)

func (s *Specs) JSON() (string, error) {
	jsonData, err := json.Marshal(&s)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

func (s *Specs) YAML() (string, error) {
	yamlData, err := yaml.Marshal(&s)
	if err != nil {
		return "", err
	}
	return string(yamlData), nil
}

func (s *Specs) TOML() (string, error) {
	tomlData, err := toml.Marshal(&s)
	if err != nil {
		return "", err
	}
	return string(tomlData), nil
}

////////////////////////////////////////////////////////////////////////////////
// CPU
////////////////////////////////////////////////////////////////////////////////

func (c CPUMaxClockSpeed) MarshalJSON() ([]byte, error) {
	return json.Marshal(float64(c) / 1e3) // already 3 decimal digits
}

// MarshalYAML Don't use yaml.Marshal for MarshalYAML()!
func (c CPUMaxClockSpeed) MarshalYAML() (any, error) {
	return float64(c) / 1e3, nil // already 3 decimal digits
}

func (c CPUMaxClockSpeed) MarshalTOML() ([]byte, error) {
	return toml.Marshal(float64(c) / 1e3) // already 3 decimal digits
}

////////////////////////////////////////////////////////////////////////////////
// L2 & L3 cache size
////////////////////////////////////////////////////////////////////////////////

// WMI returns L2 and L3 cache size in KiB

func (c L2CacheSize) MarshalJSON() ([]byte, error) {
	return json.Marshal(int64(c) / units.KiB)
}
func (c L3CacheSize) MarshalJSON() ([]byte, error) {
	return json.Marshal(int64(c) / units.KiB)
}

func (c L2CacheSize) MarshalYAML() (any, error) {
	return int64(c) / units.KiB, nil
}
func (c L3CacheSize) MarshalYAML() (any, error) {
	return int64(c) / units.KiB, nil
}

func (c L2CacheSize) MarshalTOML() ([]byte, error) {
	return toml.Marshal(int64(c) / units.KiB)
}
func (c L3CacheSize) MarshalTOML() ([]byte, error) {
	return toml.Marshal(int64(c) / units.KiB)
}

////////////////////////////////////////////////////////////////////////////////
// Memory
////////////////////////////////////////////////////////////////////////////////

func (d DIMMType) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d DIMMType) MarshalYAML() (any, error) {
	return d.String(), nil
}

func (d DIMMType) MarshalTOML() ([]byte, error) {
	return toml.Marshal(d.String())
}

// MarshalJSON Minde the int64() to prevent stack overflow
func (d DIMMCapacity) MarshalJSON() ([]byte, error) {
	return json.Marshal(int64(d) / units.GiB)
}

func (d DIMMCapacity) MarshalYAML() (any, error) {
	return int64(d) / units.GiB, nil
}

func (d DIMMCapacity) MarshalTOML() ([]byte, error) {
	return toml.Marshal(int64(d) / units.GiB)
}

////////////////////////////////////////////////////////////////////////////////
// Disk
////////////////////////////////////////////////////////////////////////////////

func (d DiskSize) MarshalJSON() ([]byte, error) {
	return json.Marshal(int64(d) / units.GB)
}

func (d DiskSize) MarshalYAML() (any, error) {
	return int64(d) / units.GB, nil
}

func (d DiskSize) MarshalTOML() ([]byte, error) {
	return toml.Marshal(int64(d) / units.GB)
}

////////////////////////////////////////////////////////////////////////////////
// Windows
////////////////////////////////////////////////////////////////////////////////

func (d WinInstallDate) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.String())
}

func (d WinInstallDate) MarshalYAML() (any, error) {
	return d.String(), nil
}

func (d WinInstallDate) MarshalTOML() ([]byte, error) {
	return toml.Marshal(d.String())
}
