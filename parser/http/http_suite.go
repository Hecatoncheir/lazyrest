package http

import nethttp "net/http"

type HttpSuite struct {
	Name         string
	Method       string
	Uri          string
	Header       nethttp.Header
	Body         string
	BodyType     string
	IsHurl       bool
	HurlFilePath string
	SecretValues []string
}

func NewHttpSuite() HttpSuite {
	return HttpSuite{
		Header: nethttp.Header{},
	}
}
