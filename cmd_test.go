package main

import (
	"testing"

	"github.com/pirmd/cli/app"
)

func TestCmdConfigExample(t *testing.T) {
	cfg := newConfig()
	cmd := newApp(cfg)

	cmd.Config.Files = []*app.ConfigFile{
		{Name: "config.example.yaml"},
	}

	if err := cmd.Config.Load(); err != nil {
		t.Fatalf("cannot read config: %s", err)
	}

	if _, err := newGostore(cfg); err != nil {
		t.Fatalf("cannot generate gostore from config: %s", err)
	}
}
