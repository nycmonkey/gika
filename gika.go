package gika

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type Client struct {
	client   *http.Client
	url      string
	splitter *regexp.Regexp
}

type FileInfo struct {
	ContentType     string `json:"Content-Type"`
	ApplicationName string `json:"Application-Name,omitempty"`
	Author          string `json:"Author,omitempty"`
}

// Parse requests the text of a file from an Apache Tika server
func (t *Client) Parse(body io.Reader, contentType string) (out []byte, err error) {
	req, err := http.NewRequest("PUT", t.url+"/tika", body)
	if err != nil {
		return
	}

	if contentType != "" {
		req.Header.Add("Content-Type", contentType)
	}

	req.Header.Add("Accept", "text/plain")

	resp, err := t.client.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return out, fmt.Errorf(resp.Status)
	}
	x := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	out, err = io.ReadAll(transform.NewReader(resp.Body, x))
	if err != nil {
		return
	}
	out = t.splitter.ReplaceAll(out, []byte("\n"))
	return
}

// GetMetadata requests metadata about a file from an Apache Tika server
func (t *Client) GetMetadata(body io.Reader, filename string) (result map[string]string, err error) {
	req, err := http.NewRequest("PUT", t.url+"/meta", body)
	if err != nil {
		return
	}

	req.Header.Add("Content-Disposition", fmt.Sprintf("attachment; filename=%s", url.QueryEscape(filename)))
	req.Header.Add("Accept", `text/csv`)

	resp, err := t.client.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf(resp.Status)
	}
	result = make(map[string]string)
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		vals := strings.Split(scanner.Text(), ",")
		if len(vals) < 2 {
			continue
		}
		k := vals[0]
		v := vals[len(vals)-1]
		if len(k) < 2 || len(v) < 2 {
			continue
		}
		result[k[1:len(k)-1]] = v[1 : len(v)-1]
	}
	err = scanner.Err()
	return
}

// DetectType requests the mime type of a file from an Apache Tika server
func (t *Client) DetectType(body io.Reader, filename string) (contentType string, err error) {
	req, err := http.NewRequest("PUT", t.url+"/detect/stream", body)
	if err != nil {
		return
	}
	req.Header.Add("Content-Disposition", fmt.Sprintf("attachment; filename=%s", url.QueryEscape(filename)))

	resp, err := t.client.Do(req)
	if err != nil {
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return contentType, fmt.Errorf(resp.Status)
	}
	var dataType []byte
	dataType, err = io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	return string(dataType), nil
}

// RecursiveParse requests the text and metadata for a container document and all embedded documents
func (t *Client) RecursiveParse(body io.Reader, contentType string) (out []byte, err error) {
	req, err := http.NewRequest("PUT", t.url+"/rmeta/text", body)
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
	out, err = io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	out = bytes.TrimSpace(t.splitter.ReplaceAll(out, []byte("\n\n")))
	return
}

func NewTika(addr string) (*Client, error) {
	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}
	u.Scheme = "http"
	u.Path = ""
	splitter := regexp.MustCompile(`(?m:\pZ*\n\pZ*\n+)`)
	return &Client{client: &http.Client{}, url: u.String(), splitter: splitter}, nil
}

func NewTikaFromDockerEnv() (*Client, error) {
	tcpAddr := os.Getenv("TIKA_PORT")
	if tcpAddr == "" {
		return nil, fmt.Errorf("'TIKA_PORT' environment variable not set; expected to find the Tika endpoint")
	}

	u, err := url.Parse(tcpAddr)
	if err != nil {
		return nil, err
	}

	u.Scheme = "http"
	return NewTika(u.String())
}
