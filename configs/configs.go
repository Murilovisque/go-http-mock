package configs

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
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
			resp := &r.Methods[i].Conversations[j].Response
			if resp.BodyPath != "" {
				resp.bodyhandler = dynamicBody{}
			} else {
				resp.bodyhandler = staticBody{}
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
	var conversations []Conversation
	pathParams := mux.Vars(r)
	for _, c := range m.Conversations {
		if hasPathParam && c.Request.hasPathParam() {
			if c.Request.matchPathParam(pathParams) {
				conversations = append(conversations, c)
			}
		} else if !hasPathParam && !c.Request.hasPathParam() {
			conversations = append(conversations, c)
		}
	}
	if len(conversations) == 0 {
		return nil
	}
	queryParams := r.URL.Query()
	if len(queryParams) == 0 {
		for _, c := range conversations {
			if len(c.Request.QueryParams) == 0 {
				return &c.Response
			}
		}
		return nil
	}

conv:
	for _, c := range conversations {
		if len(queryParams) != len(c.Request.QueryParams) {
			continue conv
		}
		for paramKey, paramVals := range queryParams {
			if !c.Request.matchQueryValues(paramKey, paramVals) {
				continue conv
			}
		}
		return &c.Response
	}
	return nil
}

type Conversation struct {
	Request  Request `json:"request,omitempty"`
	Response Response
}

type Request struct {
	PathParam   []Param            `json:"path-param,omitempty"`
	QueryParams []ParamMultiValues `json:"query-param,omitempty"`
}

func (r *Request) hasPathParam() bool {
	return len(r.PathParam) > 0
}

func (r *Request) matchPathParam(params map[string]string) bool {
	if len(r.PathParam) != len(params) {
		return false
	}
	for _, p := range r.PathParam {
		if v, ok := params[p.Name]; !ok || p.Value != v {
			return false
		}
	}
	return true
}

func (r *Request) matchQueryValues(queryParam string, valuesParam []string) bool {
	for _, v := range r.QueryParams {
		if queryParam == v.Name {
			for _, valueParam := range valuesParam {
				var matchValue bool
				for _, ValueReq := range v.Value {
					if valueParam == ValueReq {
						matchValue = true
						break
					}
				}
				if !matchValue {
					return false
				}
			}
			return true
		}
	}
	return false
}

type Param struct {
	Name  string
	Value string
}

type ParamMultiValues struct {
	Name  string
	Value []string
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
