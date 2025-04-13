package yoitsu

import (
	"io"
	"net/http"
	"net/url"
)

type Source interface {
	Json() ([]byte, error)
	Name() string
}

func NewReaderSource(name string, r io.Reader) Source {
	return &readerSource{
		r:    r,
		name: name,
	}
}

func NewUrlSource(name string, u string, opts ...Option[urlSource]) Source {
	us := urlSource{
		url:  u,
		name: name,
	}

	for _, opt := range opts {
		opt(us)
	}

	if us.httpClient == nil {
		us.httpClient = http.DefaultClient
	}

	return &us
}

type readerSource struct {
	r    io.Reader
	b    []byte
	name string
}

func (r *readerSource) Json() ([]byte, error) {
	if r.b != nil {
		return r.b, nil
	}

	var err error
	r.b, err = io.ReadAll(r.r)
	return r.b, err
}

func (r *readerSource) Name() string {
	return r.name
}

type urlSource struct {
	httpClient *http.Client
	url        string
	b          []byte
	name       string
}

func (r *urlSource) Json() ([]byte, error) {
	if r.b != nil {
		return r.b, nil
	}

	parsedUrl, err := url.Parse(r.url)
	if err != nil {
		return nil, err
	}

	r.url = parsedUrl.String()

	resp, err := r.httpClient.Get(r.url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	r.b = b
	return b, nil
}

func (r *urlSource) Name() string {
	return r.name
}

func UrlSourceWithHttpClient(c *http.Client) Option[urlSource] {
	return func(source urlSource) {
		source.httpClient = c
	}
}
