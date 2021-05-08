package persistence

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
)

type Persistence struct {
	thing interface {
		Lock()
		Unlock()
	}
	fd         *os.File
	updateChan <-chan struct{}
}

type Key = string

const (
	KeyHosts Key = "hosts"
)

func New(location string, thing interface {
	Lock()
	Unlock()
}) Persistence {

	persistFn := path.Join(location, "persist.json")

	fd, err := os.OpenFile(persistFn, os.O_CREATE|os.O_RDWR, os.ModePerm)

	if err != nil {
		panic(err)
	}

	p := Persistence{fd: fd, thing: thing}

	p.load()
	return p
}

func (p Persistence) load() {
	p.thing.Lock()

	buf, err := io.ReadAll(p.fd)
	if err != nil {
		panic(err)
	}

	//var newDat map[string]interface{}
	defer p.thing.Unlock()
	if err := json.Unmarshal(buf, &p.thing); err != nil {
		fmt.Printf("Error loading file %s", err)
		return
	}

}
func (p Persistence) Save() {
	buf, err := json.Marshal(p.thing)
	if err != nil {
		fmt.Printf("Error marshalling")
		return
	}

	p.fd.Seek(0, 1)

	writer := bufio.NewWriter(p.fd)
	writer.Write(buf)
	writer.Flush()
}

func (p Persistence) SetUpdateChan(c <-chan struct{}) {
	p.updateChan = c
	go p.waitOnUpdateChan()
}

func (p Persistence) waitOnUpdateChan() {
	for {
		select {
		case <-p.updateChan:
			fmt.Printf("Recv update hint")
			p.Save()
		}

	}

}
