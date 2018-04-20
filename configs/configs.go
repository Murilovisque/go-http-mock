package configs

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"regexp"
	"sync"
)

// GetHTTPConfig returns http-configs
func GetHTTPConfig() (*HTTPConfig, error) {
	file, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		return nil, errors.New("It was not possible to read configuration file - " + err.Error())
	}
	var config HTTPConfig
	err = json.Unmarshal(file, &config) //TODO: do validation of json
	if err != nil {
		return nil, err
	}
	for i, resource := range config.Resources {
		for j := range resource.Responses {
			config.Resources[i].Responses[j].init()
		}
	}
	return &config, nil
}

// HTTPConfig model
type HTTPConfig struct {
	Port      int
	Resources []Resource
}

// Resource model
type Resource struct {
	Name      string
	Method    string
	Path      string
	Responses []Response
	pos       int
	mutex     sync.Mutex
}

// Response model
type Response struct {
	ContentType string `json:"content-type"`
	Code        int
	Body        string `json:"body,omitempty"`
	BodyPath    string `json:"body-path,omitempty"`
	bodyhandler bodyhandler
}

func (r *Response) init() {
	if len(r.Body) > 0 {
		r.bodyhandler = staticBody{}
	} else {
		r.bodyhandler = dynamicBody{}
	}
}

// GetResponse returns the body
func (r *Resource) GetResponse() Response {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.pos >= len(r.Responses) {
		r.pos = 0
	}
	response := r.Responses[r.pos]
	r.pos++
	return response
}

// HasImageHeader check with header is image
func (r *Response) HasImageHeader() bool {
	return regexp.MustCompile("image/\\w+").MatchString(r.ContentType)
}

// GetBody returns the body
func (r *Response) GetBody() (interface{}, error) {
	return r.bodyhandler.getBody(r)
}

type bodyhandler interface {
	getBody(r *Response) (interface{}, error)
}

type staticBody struct{}

func (s staticBody) getBody(r *Response) (interface{}, error) {
	return r.Body, nil
}

type dynamicBody struct{}

func (d dynamicBody) getBody(r *Response) (interface{}, error) {
	ret, err := ioutil.ReadFile(r.BodyPath)
	if r.HasImageHeader() {
		return ret, err
	} else {
		return string(ret), err
	}
}
