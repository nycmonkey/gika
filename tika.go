package tika

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

type Tika struct {
	client *http.Client
	url    string
}

func (t *Tika) Parse(body io.Reader, contentType string) (out []byte, err error) {
	req, err := http.NewRequest("PUT", t.url, body)
	if err != nil {
		return
	}

	if contentType != "" {
		req.Header.Add("Content-Type", contentType)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return out, fmt.Errorf(resp.Status)
	}
	return ioutil.ReadAll(resp.Body)
}

func NewTika(addr string) (*Tika, error) {
	return &Tika{client: &http.Client{}, url: addr}, nil
}

func NewTikaFromDockerEnv() (*Tika, error) {
	tcpAddr := os.Getenv("TIKA_PORT")
	if tcpAddr == "" {
		return nil, fmt.Errorf("'TIKA_PORT' environment variable not set; expected to find the Tika endpoint")
	}

	u, err := url.Parse(tcpAddr)
	if err != nil {
		return nil, err
	}

	u.Scheme = "http"
	u.Path = "tika"
	return NewTika(u.String())
}
