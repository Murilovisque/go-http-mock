package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"sync"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("You need to inform the configuration file")
	}
	httpConfig, err := getHTTPConfig()
	if err != nil {
		log.Println(err)
	}
	generateMocks(httpConfig)
	log.Printf("Server mock running at port %d\n", httpConfig.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", httpConfig.Port), nil))
}

func generateMocks(httpConfig *httpConfig) {
	for _, hc := range httpConfig.Resources {
		var bodyHandle httpBody
		if len(hc.Body) > 0 {
			bodyHandle = &bodyStatic{body: hc.Body}
		} else {
			bodyHandle = &bodyFile{bodyPath: hc.BodyPath, contentType: hc.ContentType}
		}
		http.HandleFunc(hc.Path, generateHTTPHandle(bodyHandle, hc))
		log.Printf("Resource '%s' created. Method '%s'\n", hc.Path, hc.Method)
	}
}

func generateHTTPHandle(bodyHandle httpBody, hc resource) func(w http.ResponseWriter, r *http.Request) {
	b := bodyHandle
	return func(w http.ResponseWriter, r *http.Request) {
		bodyResp, err := b.get()
		if err == nil {
			if isImageHeader(hc.ContentType) {
				img, _, err := image.Decode(bytes.NewReader(bodyResp.([]byte)))
				if err != nil {
					w.Header().Set("Content-Type", "text/plain")
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprint(w, err)
				}
				buffer := new(bytes.Buffer)
				if err := jpeg.Encode(buffer, img, nil); err != nil {
					w.Header().Set("Content-Type", "text/plain")
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprint(w, err)
				}
				if _, err := w.Write(buffer.Bytes()); err != nil {
					w.Header().Set("Content-Type", "text/plain")
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprint(w, err)
				}
			} else {
				fmt.Fprint(w, bodyResp)
			}
			w.Header().Set("Content-Type", hc.ContentType)
			w.WriteHeader(hc.Code)
		} else {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, err)
		}
	}
}

func getHTTPConfig() (*httpConfig, error) {
	file, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		return nil, errors.New("It was not possible to read configuration file - " + err.Error())
	}
	var config httpConfig
	return &config, json.Unmarshal(file, &config) //TODO: do validation of json
}

type httpBody interface {
	get() (interface{}, error)
}

type bodyStatic struct {
	body  []string
	pos   int
	mutex sync.Mutex
}

func (b *bodyStatic) get() (interface{}, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	log.Println(b, len(b.body))
	if b.pos >= len(b.body) {
		b.pos = 0
	}
	ret := b.body[b.pos]
	b.pos++
	return ret, nil
}

type bodyFile struct {
	bodyPath    []string
	pos         int
	mutex       sync.Mutex
	contentType string
}

func (b *bodyFile) get() (interface{}, error) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	if b.pos >= len(b.bodyPath) {
		b.pos = 0
	}
	ret, err := ioutil.ReadFile(b.bodyPath[b.pos])
	b.pos++
	if isImageHeader(b.contentType) {
		return ret, err
	} else {
		return string(ret), err
	}
}

func isImageHeader(h string) bool {
	return regexp.MustCompile("image/\\w+").MatchString(h)
}

type httpConfig struct {
	Port      int
	Resources []resource
}

type resource struct {
	Name        string
	Method      string
	Path        string
	ContentType string `json:"content-type"`
	Code        int
	Body        []string `json:"body,omitempty"`
	BodyPath    []string `json:"path-body,omitempty"`
}
