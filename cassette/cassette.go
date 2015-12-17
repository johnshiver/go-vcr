// Copyright (c) 2015 Marin Atanasov Nikolov <dnaeon@gmail.com>
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions
// are met:
// 1. Redistributions of source code must retain the above copyright
//    notice, this list of conditions and the following disclaimer
//    in this position and unchanged.
// 2. Redistributions in binary form must reproduce the above copyright
//    notice, this list of conditions and the following disclaimer in the
//    documentation and/or other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE AUTHOR(S) ``AS IS'' AND ANY EXPRESS OR
// IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES
// OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED.
// IN NO EVENT SHALL THE AUTHOR(S) BE LIABLE FOR ANY DIRECT, INDIRECT,
// INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT
// NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF
// THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package cassette

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// Cassette format versions
const (
	cassetteFormatV1 = 1
)

var (
	InteractionNotFound = errors.New("Requested interaction not found")
)

// Client request type
type Request struct {
	// Body of request
	Body string `yaml:"body"`

	// Request headers
	Headers http.Header `yaml:"headers"`

	// Request URL
	URL string `yaml:"url"`

	// Request method
	Method string `yaml:"method"`
}

// Server response type
type Response struct {
	// Body of response
	Body string `yaml:"body"`

	// Response headers
	Headers http.Header `yaml:"headers"`

	// Response status message
	Status string `yaml:"status"`

	// Response status code
	Code int `yaml:"code"`
}

// Interaction type contains a pair of request/response for a
// single HTTP interaction between a client and a server
type Interaction struct {
	Request  `yaml:"request"`
	Response `yaml:"response"`
}

// Cassette type
type Cassette struct {
	// Name of the cassette
	Name string `yaml:"-"`

	// File name of the cassette as written on disk
	File string `yaml:"-"`

	// Cassette format version
	Version int `yaml:"version"`

	// Interactions between client and server
	Interactions []*Interaction `yaml:"interactions"`
}

// Creates a new empty cassette
func New(name string) *Cassette {
	c := &Cassette{
		Name:         name,
		File:         fmt.Sprintf("%s.yaml", name),
		Version:      cassetteFormatV1,
		Interactions: make([]*Interaction, 0),
	}

	return c
}

// Loads a cassette file from disk
func Load(name string) (*Cassette, error) {
	c := New(name)
	data, err := ioutil.ReadFile(c.File)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(data, &c)

	return c, err
}

// Adds a new interaction to the cassette
func (c *Cassette) AddInteraction(i *Interaction) {
	c.Interactions = append(c.Interactions, i)
}

// Gets a recorded interaction
func (c *Cassette) GetInteraction(r *http.Request) (*Interaction, error) {
	for _, i := range c.Interactions {
		if r.Method == i.Request.Method && r.URL.String() == i.Request.URL {
			return i, nil
		}
	}

	return nil, InteractionNotFound
}

// Saves the cassette on disk for future re-use
func (c *Cassette) Save() error {
	// Save cassette file only if there were any interactions made
	if len(c.Interactions) == 0 {
		return nil
	}

	// Create directory for cassette if missing
	cassetteDir := filepath.Dir(c.File)
	if _, err := os.Stat(cassetteDir); os.IsNotExist(err) {
		if err = os.MkdirAll(cassetteDir, 0755); err != nil {
			return err
		}
	}

	// Marshal to YAML and save interactions
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	f, err := os.Create(c.File)
	if err != nil {
		return err
	}

	// Honor the YAML structure specification
	// http://www.yaml.org/spec/1.2/spec.html#id2760395
	_, err = f.Write([]byte("---\n"))
	if err != nil {
		return err
	}

	_, err = f.Write(data)
	if err != nil {
		return err
	}

	return nil
}
