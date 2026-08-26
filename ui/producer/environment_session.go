package producer

import (
	"fmt"
	"net/http/cookiejar"
)

// ResetEnvironmentSession removes state that must not cross environment
// boundaries. It preserves the configured client's transport and redirect
// behavior while replacing its cookie jar.
func (widget *Producer) ResetEnvironmentSession() error {
	if widget == nil {
		return nil
	}
	widget.ClearCapturedResponses()

	widget.runnerConfigMutex.Lock()
	defer widget.runnerConfigMutex.Unlock()
	hasCookies := widget.runnerConfig.Jar != nil ||
		(widget.runnerConfig.Client != nil && widget.runnerConfig.Client.Jar != nil)
	if !hasCookies {
		return nil
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		return fmt.Errorf("reset cookie jar: %w", err)
	}
	widget.runnerConfig.Jar = jar
	if widget.runnerConfig.Client != nil {
		client := *widget.runnerConfig.Client
		client.Jar = jar
		widget.runnerConfig.Client = &client
	}
	return nil
}
