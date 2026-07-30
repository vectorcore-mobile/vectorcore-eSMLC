package config

import (
	"testing"
	"time"
)

func TestValidationRejectsUnsafeAndUnsupportedSettings(t *testing.T) {
	cases := []func(*Config){
		func(c *Config) { c.SLs.Port = 0 },
		func(c *Config) { c.SLs.ExpectedPPID = 1 },
		func(c *Config) { c.SLs.MaxMessageSize = 1 },
		func(c *Config) { c.Positioning.Method = "otdoa" },
		func(c *Config) { c.Positioning.ECID.Enabled = true },
		func(c *Config) { c.Positioning.ECID.CellDataFile = "cells.yaml" },
		func(c *Config) { c.Positioning.Simulation.ResponseDelay = time.Second },
		func(c *Config) { c.Positioning.Simulation.Enabled = true; c.Positioning.Simulation.Latitude = 91 },
	}
	for _, change := range cases {
		c := Default()
		change(&c)
		if err := c.Validate(); err == nil {
			t.Fatal("invalid configuration accepted")
		}
	}
}
