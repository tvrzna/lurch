package main

import (
	"os"
	"strconv"
	"strings"
)

type Project struct {
	name string
	dir  string
}

// Get last build number
func (p *Project) LastCount() int {
	data, err := os.ReadFile(p.dir + "/counter")
	if err != nil {
		return 0
	}
	i, _ := strconv.Atoi(strings.TrimSpace(string(data)))
	return i
}

// Rotate count
func (p *Project) RotateCount() (int, error) {
	count := p.LastCount() + 1
	return count, os.WriteFile(p.dir+"/counter", []byte(strconv.Itoa(count)), 0600)
}

// Create new build, rotate last build number and make directory for new build
func (p *Project) NewBuild() (*Build, error) {
	buildNo, err := p.RotateCount()
	if err != nil {
		return nil, err
	}
	strBuildNo := strconv.Itoa(buildNo)
	b := &Build{name: strBuildNo, dir: p.dir + "/" + strBuildNo, p: p}
	if err := os.MkdirAll(b.dir, 0755); err != nil {
		return nil, err
	}
	b.SetStatus(Unknown)
	b.c = make(chan bool)
	return b, nil
}
