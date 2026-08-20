package http

type HttpSuite struct {
	Name     string
	Method   string
	Uri      string
	Header   map[string]string
	Body     string
	BodyType string
}

func NewHttpSuite() HttpSuite {
	return HttpSuite{
		Header: map[string]string{},
	}
}
