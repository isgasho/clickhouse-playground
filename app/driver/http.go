package driver

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type httpDriver struct {
	endpoint *url.URL
}

func NewHTTPDriver(endpoint *url.URL) Driver {
	return &httpDriver{endpoint: endpoint}
}

func (d *httpDriver) HealthCheck() error {
	response, err := http.Get(d.endpoint.String())
	if err != nil {
		return err
	}

	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if !bytes.Equal(contents, []byte("Ok.\n")) {
		return fmt.Errorf("strings \"%s\", \"%s\"dont match", contents, "Ok.")
	}

	return nil
}

func (d *httpDriver) Exec(query string) (string, error) {
	response, err := http.Post(
		d.endpoint.String(),
		"application/json",
		strings.NewReader(query),
	)
	if err != nil {
		return "", err
	}
	defer func() {
		_ = response.Body.Close()
	}()

	data, err := ioutil.ReadAll(response.Body)
	if response.StatusCode != 200 {
		return "", errors.New(string(data))
	}

	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (d *httpDriver) Close() error {
	return nil
}


