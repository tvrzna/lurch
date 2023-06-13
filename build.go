package main

import (
	"bytes"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"time"
)

type BuildStatus byte

const (
	Unknown BuildStatus = iota
	Finished
	Stopped
	Failed
	InProgress
)

// Stringify build status
func (b BuildStatus) String() string {
	return []string{"unknown", "finished", "stopped", "failed", "inprogress"}[int(b)]
}

// Marshal build status type into string
func (b BuildStatus) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("\"")
	buffer.WriteString(b.String())
	buffer.WriteString("\"")

	return buffer.Bytes(), nil
}

// Unmarshal string build status into type
func (b *BuildStatus) UnmarshalJSON(data []byte) error {
	var v string
	json.Unmarshal(data, &v)

	*b = map[string]BuildStatus{
		"unknown":    Unknown,
		"finished":   Finished,
		"stopped":    Stopped,
		"failed":     Failed,
		"inprogress": InProgress,
	}[v]

	return nil
}

type Build struct {
	name string
	dir  string
	p    *Project
	c    chan bool
}

// Get status of build
func (b *Build) Status() BuildStatus {
	data, err := os.ReadFile(b.dir + "/status")
	if err != nil {
		return Unknown
	}
	i, _ := strconv.Atoi(strings.TrimSpace(string(data)))
	return BuildStatus(byte(i))
}

// Set status of build
func (b *Build) SetStatus(s BuildStatus) error {
	return os.WriteFile(b.dir+"/status", []byte(strconv.Itoa(int(s))), 0600)
}

// Get path to console ouput of build
func (b *Build) OutputPath() string {
	return b.dir + "/" + "console.log"
}

// Read content of console output of build
func (b *Build) ReadOutput() (string, error) {
	data, err := os.ReadFile(b.OutputPath())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Get start date of build
func (b *Build) StartDate() time.Time {
	s, err := os.Stat(b.dir)
	if err != nil {
		return time.UnixMicro(0)
	}
	return s.ModTime()
}

// Get end date of build
func (b *Build) EndDate() time.Time {
	s, err := os.Stat(b.dir + "/status")
	if err != nil {
		return time.UnixMicro(0)
	}
	return s.ModTime()
}

// Path to workspace
func (b *Build) WorkspacePath() string {
	return b.dir + "/workspace"
}

// Make workspace directory
func (b *Build) MkWorkspace() error {
	return os.MkdirAll(b.WorkspacePath(), 0755)
}
