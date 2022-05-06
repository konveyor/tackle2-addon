package main

import (
	"github.com/konveyor/tackle2-addon/command"
	"github.com/konveyor/tackle2-addon/nas"
	hub "github.com/konveyor/tackle2-hub/addon"
	"github.com/konveyor/tackle2-hub/api"
	"os"
	pathlib "path"
	"strings"
)

var (
	addon = hub.Addon
)

const (
	MountRoot = "/mnt/"
)

type SoftError = hub.SoftError

func main() {
	addon.Run(func() (err error) {
		variant := addon.Variant()
		addon.Activity("Variant: %s", variant)
		switch variant {
		case "mount:report":
			err = mountReport()
		case "mount:clean": // volume:report  volume:clean
			err = mountClean()
		default:
			err = &SoftError{Reason: "Variant not supported."}
		}
		return
	})
}

//
// mountReport reports mount statistics.
func mountReport() (err error) {
	d := &Mount{}
	err = addon.DataWith(d)
	if err != nil {
		err = &SoftError{Reason: err.Error()}
		return
	}
	if len(d.Names) == 0 {
		return
	}
	for _, name := range d.Names {
		var v *api.Volume
		v, err = addon.Volume.Find(name)
		if err != nil {
			return
		}
		path := MountRoot + name
		cmd := command.Command{Path: "/usr/bin/df"}
		cmd.Options.Add("-h")
		cmd.Options.Addf(path)
		err = cmd.Run()
		if err != nil {
			return
		}
		output := string(cmd.Output)
		output = strings.Split(output, "\n")[1]
		part := strings.Fields(output)
		v.Capacity = part[1]
		v.Used = part[2]
		err = addon.Volume.Update(v)
		if err != nil {
			return
		}
	}
	return
}

//
// mountClean deletes the content of the mount.
func mountClean() (err error) {
	d := &Mount{}
	err = addon.DataWith(d)
	if err != nil {
		err = &SoftError{Reason: err.Error()}
		return
	}
	if len(d.Names) == 0 {
		return
	}
	var entries []os.DirEntry
	for _, name := range d.Names {
		path := MountRoot + name
		entries, err = os.ReadDir(path)
		if err != nil {
			err = &SoftError{Reason: err.Error()}
			return
		}
		for _, entry := range entries {
			p := pathlib.Join(path, entry.Name())
			err = nas.RmDir(p)
			if err != nil {
				err = &SoftError{Reason: err.Error()}
				return
			}
		}
	}
	return
}

//
// Mount input.
type Mount struct {
	Names []string `json:"names"`
}
