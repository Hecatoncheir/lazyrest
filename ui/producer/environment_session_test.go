package producer

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"testing"

	parserhttp "github.com/Hecatoncheir/lazyrest/parser/http"
	"github.com/Hecatoncheir/lazyrest/runner"
)

func TestResetEnvironmentSessionReplacesCookiesAndKeepsClientSettings(t *testing.T) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	target, err := url.Parse("https://example.test")
	if err != nil {
		t.Fatal(err)
	}
	jar.SetCookies(target, []*http.Cookie{{Name: "session", Value: "old"}})
	transport := &http.Transport{}
	client := &http.Client{Jar: jar, Transport: transport}
	widget := &Producer{runnerConfig: runner.Config{Jar: jar, Client: client}}
	widget.recordResponse(parserhttp.HttpSuite{Name: "login", SourceFilePath: "requests.http"}, runner.Response{Code: "200 OK", Body: "secret"}, nil)

	if err := widget.ResetEnvironmentSession(); err != nil {
		t.Fatal(err)
	}
	config := widget.runnerConfiguration()
	if config.Client == client || config.Client.Transport != transport {
		t.Fatal("HTTP client settings were not preserved in a new client value")
	}
	if config.Jar == jar || config.Client.Jar != config.Jar || len(config.Jar.Cookies(target)) != 0 {
		t.Fatal("cookie session was not replaced with an empty jar")
	}
	if captures := widget.CapturedResponses(); len(captures) != 0 {
		t.Fatalf("captured responses crossed environment boundary: %+v", captures)
	}
}
