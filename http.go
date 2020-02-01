package main

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

//HTTPGateway struct
type HTTPGateway struct {
	Client *http.Client
}

//Get func
func (hg *HTTPGateway) Get(resourceURL string) (int, io.Reader, error) {
	response, err := hg.Client.Get(resourceURL)
	if err != nil {
		return 0, nil, errors.WithStack(err)
	}
	defer response.Body.Close()
	request, err := http.NewRequest("GET", resourceURL, nil)
	if err != nil {
		return 0, nil, errors.WithStack(err)
	}
	logErr := logRequest(request)
	if logErr != nil {
		return 0, nil, errors.WithStack(logErr)
	}
	response, resErr := hg.Client.Do(request)
	if resErr != nil {
		return 0, nil, errors.WithStack(resErr)
	}
	resBytes, logResErr := logResponse(response)
	if logResErr != nil {
		return 0, nil, errors.WithStack(logResErr)
	}
	return response.StatusCode, bytes.NewBuffer(resBytes), nil
}

func logRequest(request *http.Request) error {
	logStringBuilder := new(strings.Builder)
	logStringBuilder.WriteString("Http request:\n")
	logStringBuilder.WriteString("Method ")
	logStringBuilder.WriteString(request.Method)
	logStringBuilder.WriteString(" ")
	logStringBuilder.WriteString(request.URL.String())
	logStringBuilder.WriteString("\n")
	log.Println(logStringBuilder)
	return nil
}

func logResponse(response *http.Response) ([]byte, error) {
	logStringBuilder := new(strings.Builder)
	logStringBuilder.WriteString("Http response:\n")
	logStringBuilder.WriteString("Http status: ")
	logStringBuilder.WriteString(strconv.Itoa(response.StatusCode))
	logStringBuilder.WriteString("\n")
	data, readErr := ioutil.ReadAll(response.Body)
	if readErr != nil {
		return nil, errors.WithStack(readErr)
	}
	logStringBuilder.Write(data)
	logStringBuilder.WriteString("\n")
	log.Print(logStringBuilder)
	return data, nil
}
