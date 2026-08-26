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
	// HurlEntry is the position of an entry in its Hurl file, counted from
	// one. Hurl runs a file in order, so an entry is reached by running
	// everything up to it.
	HurlEntry int
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
