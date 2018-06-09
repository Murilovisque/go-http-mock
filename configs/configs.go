package configs

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/gorilla/mux"
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
	log.Printf("%#v", config)
	for _, resource := range config.Resources {
		resource.init()
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
	Path    string
	Methods []Method
}

func (r *Resource) init() {
	for i := range r.Methods {
		r.Methods[i].Type = strings.ToUpper(r.Methods[i].Type)
		for j := range r.Methods[i].Conversations {
			resp := r.Methods[i].Conversations[j].Response
			if resp.Body != "" {
				resp.bodyhandler = staticBody{}
			} else {
				resp.bodyhandler = dynamicBody{}
			}
		}
	}
}

// Response returns the body
func (r *Resource) Response(req *http.Request) *Response {
	for _, m := range r.Methods {
		if m.Type == req.Method {
			return m.Response(req, r.hasPathParam())
		}
	}
	return nil
}

func (r *Resource) hasPathParam() bool {
	return regexp.MustCompile(`\/[\w\-\/]*\{[\w\-]+\}([\w\-\/]|\{[\w\-]+\})*`).MatchString(r.Path)
}

type Method struct {
	Name          string
	Type          string
	Conversations []Conversation
}

func (m Method) String() string {
	return fmt.Sprintf("Method: %s - Name: %s", m.Type, m.Name)
}

func (m *Method) Response(r *http.Request, hasPathParam bool) *Response {
	params := mux.Vars(r)
	for _, c := range m.Conversations {
		if hasPathParam && c.Request != nil && c.Request.ParamPath != nil {
			p := c.Request.ParamPath
			if val, ok := params[p.Name]; ok && val == p.Value {
				return c.Response
			}
		} else if !hasPathParam && (c.Request == nil || c.Request.ParamPath == nil) {
			return c.Response
		}
	}
	return nil
}

type Conversation struct {
	Request  *Request
	Response *Response
}

type Request struct {
	ParamPath *Param `json:"param-path"`
}

type Param struct {
	Name  string
	Value string
}

// Response model
type Response struct {
	ContentType string `json:"content-type"`
	Code        int
	Body        string `json:"body,omitempty"`
	BodyPath    string `json:"body-path,omitempty"`
	bodyhandler bodyhandler
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
	}
	return string(ret), err
}
