package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Project struct {
	name   string
	dir    string
	params map[string]string
}

// Get last job number
func (p *Project) LastCount() int {
	data, err := os.ReadFile(filepath.Join(p.dir, "counter"))
	if err != nil {
		return 0
	}
	i, _ := strconv.Atoi(strings.TrimSpace(string(data)))
	return i
}

// Rotate count
func (p *Project) RotateCount() (int, error) {
	count := p.LastCount() + 1
	return count, os.WriteFile(filepath.Join(p.dir, "counter"), []byte(strconv.Itoa(count)), 0600)
}

// Create new job, rotate last job number and make directory for new job
func (p *Project) NewJob() (*Job, error) {
	jobNo, err := p.RotateCount()
	if err != nil {
		return nil, err
	}
	strJobNo := strconv.Itoa(jobNo)
	b := &Job{name: strJobNo, dir: filepath.Join(p.dir, strJobNo), p: p}
	if err := os.MkdirAll(b.dir, 0755); err != nil {
		return nil, err
	}
	b.SetStatus(Unknown)
	b.interrupt = make(chan bool)
	return b, nil
}

// Sets params in expected format, exclude all params with incorrect format
func (p *Project) SetParams(params map[string]string) {
	if checkedParams := checkParams(params); checkedParams != nil {
		p.params = checkedParams
	}
}

// Saves params into file
func (p *Project) SaveParams() error {
	return saveParams(filepath.Join(p.dir, "params"), p.params)
}

// Loads params from file into map, if not found, leave method without drama
func (p *Project) LoadParams() {
	p.params = loadParams(filepath.Join(p.dir, "params"))
}
