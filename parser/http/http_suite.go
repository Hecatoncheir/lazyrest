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
	// Variables are passed to Hurl, which does its own substitution.
	Variables map[string]string
	// GraphQLVariables and GraphQLOperation complete a GraphQL request whose
	// query is held in Body.
	GraphQLVariables string
	GraphQLOperation string
	SecretValues     []string
}

func NewHttpSuite() HttpSuite {
	return HttpSuite{
		Header: nethttp.Header{},
	}
}
