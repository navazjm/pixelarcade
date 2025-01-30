package webapp

import (
	"testing"
)

func TestNewApplication(t *testing.T) {
	app := New()

	if app.Logger == nil {
		t.Errorf("expected logger to be initialized, got nil")
	}

	if app.Config == nil {
		t.Errorf("expected config to be initialized, got nil")
	}
}

// TODO:
func TestApplicationInitServices(t *testing.T) {}
