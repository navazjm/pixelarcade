package webapp

import (
	"testing"
)

func TestNewApplicationAndInitServices(t *testing.T) {
	app := New()

	if app.Logger == nil {
		t.Errorf("expected logger to be initialized, got nil")
	}

	if app.Config == nil {
		t.Errorf("expected config to be initialized, got nil")
	}

	app.InitServices(nil)
	if app.AuthService == nil {
		t.Errorf("expected Auth service to be initialized, got nil")
	}
	if app.GamesService == nil {
		t.Errorf("expected Games service to be initialized, got nil")
	}
}
