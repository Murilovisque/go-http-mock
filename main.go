package main

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"github.com/Murilovisque/go-http-mock/configs"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("You need to inform the configuration file")
	}
	httpConfig, err := configs.GetHTTPConfig()
	if err != nil {
		log.Println(err)
	}
	generateMocks(httpConfig)
	log.Printf("Server mock running at port %d\n", httpConfig.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", httpConfig.Port), nil))
}

func generateMocks(httpConfig *configs.HTTPConfig) {
	r := mux.NewRouter()
	for _, hc := range httpConfig.Resources {
		r.HandleFunc(hc.Path, generateHTTPHandle(hc))
		log.Printf("Resource '%s' created. Methods '%s'\n", hc.Path, hc.Methods)
	}
	http.Handle("/", r)
}

func generateHTTPHandle(resource configs.Resource) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		response := resource.Response(r)
		if response == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		body, err := response.GetBody()
		if err != nil {
			returnInternalError(w, err)
		} else if response.HasImageHeader() {
			returnImage(w, body, response)
		} else {
			w.WriteHeader(response.Code)
			w.Header().Set("Content-Type", response.ContentType)
			fmt.Fprint(w, body)
		}
	}
}

func returnInternalError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, err)
}

func returnImage(w http.ResponseWriter, body interface{}, configResp *configs.Response) {
	img, _, err := image.Decode(bytes.NewReader(body.([]byte)))
	if err != nil {
		returnInternalError(w, err)
	}
	buffer := new(bytes.Buffer)
	if err := jpeg.Encode(buffer, img, nil); err != nil {
		returnInternalError(w, err)
	}
	if _, err := w.Write(buffer.Bytes()); err != nil {
		returnInternalError(w, err)
	}
	w.WriteHeader(configResp.Code)
	w.Header().Set("Content-Type", configResp.ContentType)
}
