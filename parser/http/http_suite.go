package http

type HttpSuite struct {
	method   string
	uri      string
	header   map[string]string
	body     string
	bodyType string
}

func NewHttpSuite() HttpSuite {
	return HttpSuite{
		header: map[string]string{},
	}
}

func (suite *HttpSuite) Method() string {
	return suite.method
}

func (suite *HttpSuite) Uri() string {
	return suite.uri
}

func (suite *HttpSuite) Header() map[string]string {
	return suite.header
}

func (suite *HttpSuite) Body() string {
	return suite.body
}

func (suite *HttpSuite) BodyType() string {
	return suite.bodyType
}
