package main

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type Context struct {
	mutex  *sync.Mutex
	port   int
	path   string
	appUrl string
	builds []*Build
}

// Init new build context
func NewContext(port int, path string) *Context {
	var mutex sync.Mutex

	absPath, err := filepath.Abs(path)
	if err != nil {
		log.Fatal("Unknown path")
	}

	return &Context{port: port, path: absPath, builds: make([]*Build, 0), mutex: &mutex}
}

func (c *Context) getAppUrl() string {
	if c.appUrl == "" {
		return "http://" + c.getServerUri()
	}
	return c.appUrl
}

func (c *Context) getServerUri() string {
	return "localhost:" + strconv.Itoa(c.port)
}

func (c *Context) Build(p *Project) bool {
	// Check if being build
	if c.isProjectBeingBuilt(p) {
		return false
	}

	b, err := p.NewBuild()
	if err != nil {
		return false
	}

	c.mutex.Lock()
	c.builds = append(c.builds, b)
	c.mutex.Unlock()

	c.removeOldBuilds(p)

	go c.build(b)
	return true
}

func (c *Context) Interrupt(b *Build) {
	c.mutex.Lock()
	for _, build := range c.builds {
		if b.p.name == build.p.name && b.name == build.name {
			log.Printf("-- interrupting build #%s of %s", b.name, b.p.name)
			build.c <- true
		}
	}
	defer c.mutex.Unlock()
}

func (c *Context) build(b *Build) {
	log.Printf(">> started build #%s for %s", b.name, b.p.name)
	cmd := exec.Command("sh", "-c", b.p.dir+"/script.sh")
	b.MkWorkspace()
	cmd.Dir = b.WorkspacePath()
	output, err := os.OpenFile(b.OutputPath(), os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		b.SetStatus(Failed)
		c.removeFromSlice(b)
		log.Printf("-- failed to open output for #%s of %s", b.name, b.p.name)
		return
	}
	defer output.Close()
	cmd.Stdin = output
	cmd.Stdout = output

	cmd.Start()
	go c.watchForInterrupt(b, cmd)

	if err := cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if exiterr.ExitCode() == 0 {
				b.SetStatus(Finished)
			} else if b.Status() != Stopped {
				b.SetStatus(Failed)
			}
			output.WriteString(exiterr.String())
		} else {
			output.WriteString("\nFailed!")
			b.SetStatus(Failed)
		}
	} else {
		b.SetStatus(Finished)
	}
	close(b.c)
	c.removeFromSlice(b)

	if err := c.compressFolder(b.WorkspacePath()+".tar.gz", b.WorkspacePath()); err != nil {
		log.Print("-- could not compress", b.WorkspacePath())
	}
	os.RemoveAll(b.WorkspacePath())

	log.Printf("<< finished build #%s for %s", b.name, b.p.name)
}

func (c *Context) removeFromSlice(b *Build) {
	c.mutex.Lock()
	if index := c.indexOf(b); index >= 0 {
		if len(c.builds) == 1 {
			c.builds = c.builds[:0]
		} else {
			c.builds = append(c.builds[:index], c.builds[index+1])
		}
	}
	c.mutex.Unlock()
}

func (c *Context) indexOf(b *Build) int {
	for i, build := range c.builds {
		if b.p.name == build.p.name && b.name == build.name {
			return i
		}
	}
	return -1
}

func (c *Context) watchForInterrupt(b *Build, cmd *exec.Cmd) {
	status := <-b.c
	if status {
		b.SetStatus(Stopped)
		cmd.Process.Signal(os.Kill)
	}
}

func (c *Context) IsBeingBuilt(b *Build) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.indexOf(b) >= 0
}

func (c *Context) isProjectBeingBuilt(p *Project) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for _, b := range c.builds {
		if b.p.name == p.name {
			return true
		}
	}
	return false
}

func (c *Context) removeOldBuilds(p *Project) {
	builds, _ := c.ListBuilds(p)
	if len(builds) > 10 {
		builds = builds[10:]
		log.Print("-- remove old builds of ", p.name)
		for _, b := range builds {
			os.RemoveAll(b.dir)
		}
	}
}

func (c *Context) ListProjects() ([]*Project, error) {
	result := make([]*Project, 0)
	entries, err := os.ReadDir(c.path)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() {
			result = append(result, &Project{name: e.Name(), dir: c.path + "/" + e.Name()})
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return strings.Compare(result[i].name, result[j].name) <= 0
	})

	return result, nil
}

func (c *Context) OpenProject(name string) *Project {
	return &Project{name: name, dir: c.path + "/" + name}
}

func (c *Context) ListBuilds(p *Project) ([]*Build, error) {
	result := make([]*Build, 0)
	entries, err := os.ReadDir(p.dir)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() {
			result = append(result, &Build{name: e.Name(), dir: p.dir + "/" + e.Name(), p: p})
		}
	}

	sort.Slice(result, func(i, j int) bool {
		bI, _ := strconv.Atoi(result[i].name)
		bJ, _ := strconv.Atoi(result[j].name)
		return bI > bJ
	})

	return result, nil
}

func (c *Context) OpenBuild(p *Project, name string) *Build {
	return &Build{name: name, dir: p.dir + "/" + name, p: p}
}

func (c *Context) compressFolder(outputPath, inputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	zw := gzip.NewWriter(file)
	tw := tar.NewWriter(zw)

	err = filepath.Walk(inputPath, func(file string, fi os.FileInfo, e error) error {
		header, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return err
		}

		header.Name = strings.Replace(filepath.ToSlash(file), inputPath, "", 1)
		if header.Name == "" {
			return nil
		}

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !fi.IsDir() {
			data, err := os.Open(file)
			if err != nil {
				return err
			}
			if _, err := io.Copy(tw, data); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	if err := tw.Close(); err != nil {
		return err
	}

	if err := zw.Close(); err != nil {
		return err
	}
	return nil
}
