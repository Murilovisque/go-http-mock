package configs

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"regexp"
	"sync"
	"fmt"
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
		for j := range resource.Methods {
			for k := range resource.Methods[j].Responses {
				config.Resources[i].Methods[j].Responses[k].init()
			}
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
	Path      string
	Methods []Method
}

type Method struct {
	Name      string
	Method    string
	Responses []Response
	pos       int
	mutex     sync.Mutex
}

func (m Method) String() string {
    return fmt.Sprintf(m.Method)
}

// Response model
type Response struct {
	ContentType string `json:"content-type"`
	Code        int
	Parameter   bool
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
func (r *Method) GetResponse(hasParameter bool) Response {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	for _, response := range r.Responses {
		if response.Parameter == hasParameter {
			return response
		}
	}
	return r.Responses[0]
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
