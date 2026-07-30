package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"net"
	"os"
	"time"
)

type Config struct {
	Service     Service     `yaml:"service"`
	SLs         SLs         `yaml:"sls"`
	Positioning Positioning `yaml:"positioning"`
}
type Service struct {
	Name     string `yaml:"name"`
	LogLevel string `yaml:"log_level"`
}
type SLs struct {
	Enabled         bool          `yaml:"enabled"`
	ListenAddress   string        `yaml:"listen_address"`
	Port            int           `yaml:"port"`
	ExpectedPPID    uint32        `yaml:"expected_ppid"`
	MaxAssociations int           `yaml:"max_associations"`
	MaxSessions     int           `yaml:"max_sessions"`
	SessionTimeout  time.Duration `yaml:"session_timeout"`
	MaxMessageSize  int           `yaml:"max_message_size"`
}
type Positioning struct {
	Method     string     `yaml:"method"`
	ECID       ECIDPolicy `yaml:"ecid"`
	Simulation Simulation `yaml:"simulation"`
}

// ECIDPolicy is local operator authorization only. It never asserts UE
// capability; that is discovered through LPP ProvideCapabilities.
type ECIDPolicy struct {
	Enabled               bool          `yaml:"enabled"`
	RequestRSRP           bool          `yaml:"request_rsrp"`
	RequestRSRQ           bool          `yaml:"request_rsrq"`
	RequestUERxTxTimeDiff bool          `yaml:"request_ue_rxtx_time_diff"`
	CellDataFile          string        `yaml:"cell_data_file"`
	CellDataMaxAge        time.Duration `yaml:"cell_data_max_age"`
}
type Simulation struct {
	Enabled           bool          `yaml:"enabled"`
	Latitude          float64       `yaml:"latitude"`
	Longitude         float64       `yaml:"longitude"`
	UncertaintyMeters float64       `yaml:"uncertainty_meters"`
	ResponseDelay     time.Duration `yaml:"response_delay"`
	FailureCause      *uint8        `yaml:"failure_cause"`
}

func Default() Config {
	return Config{Service: Service{Name: "vectorcore-esmlc", LogLevel: "info"}, SLs: SLs{Enabled: true, ListenAddress: "0.0.0.0", Port: 9082, ExpectedPPID: 29, MaxAssociations: 32, MaxSessions: 10000, SessionTimeout: 10 * time.Second, MaxMessageSize: 1 << 20}, Positioning: Positioning{Method: "gnss", Simulation: Simulation{Enabled: false}}}
}
func Load(path string) (Config, error) {
	c := Default()
	b, e := os.ReadFile(path)
	if e != nil {
		return c, e
	}
	if e = yaml.Unmarshal(b, &c); e != nil {
		return c, fmt.Errorf("config: parse YAML: %w", e)
	}
	return c, c.Validate()
}
func (c Config) Validate() error {
	if c.Service.Name == "" {
		return fmt.Errorf("config: service.name is required")
	}
	if c.Service.LogLevel != "" && c.Service.LogLevel != "debug" && c.Service.LogLevel != "info" && c.Service.LogLevel != "warn" && c.Service.LogLevel != "error" {
		return fmt.Errorf("config: invalid log level")
	}
	s := c.SLs
	if !s.Enabled {
		return nil
	}
	if net.ParseIP(s.ListenAddress) == nil {
		return fmt.Errorf("config: sls.listen_address must be an IP address")
	}
	if s.Port < 1 || s.Port > 65535 {
		return fmt.Errorf("config: sls.port must be 1..65535")
	}
	if s.ExpectedPPID != 29 {
		return fmt.Errorf("config: expected PPID must be 29")
	}
	if s.MaxAssociations < 1 || s.MaxSessions < 1 {
		return fmt.Errorf("config: association and session limits must be positive")
	}
	if s.SessionTimeout <= 0 {
		return fmt.Errorf("config: session timeout must be positive")
	}
	if s.MaxMessageSize < 1024 || s.MaxMessageSize > aperLimit {
		return fmt.Errorf("config: max message size must be 1024..1048576")
	}
	p := c.Positioning
	if p.Method != "gnss" {
		return fmt.Errorf("config: unsupported positioning method")
	}
	x := p.Simulation
	if p.ECID.Enabled && !p.ECID.RequestRSRP && !p.ECID.RequestRSRQ && !p.ECID.RequestUERxTxTimeDiff {
		return fmt.Errorf("config: enabled ECID requires at least one requested measurement")
	}
	if p.ECID.CellDataFile != "" && p.ECID.CellDataMaxAge <= 0 {
		return fmt.Errorf("config: ECID cell data requires a positive maximum age")
	}
	if p.ECID.CellDataFile == "" && p.ECID.CellDataMaxAge != 0 {
		return fmt.Errorf("config: ECID cell data maximum age requires a file")
	}
	if x.ResponseDelay < 0 {
		return fmt.Errorf("config: simulation response delay must not be negative")
	}
	if x.ResponseDelay != 0 {
		return fmt.Errorf("config: simulation response delay is unsupported by the job-scoped estimator")
	}
	if x.Enabled && (x.Latitude < -90 || x.Latitude > 90 || x.Longitude < -180 || x.Longitude > 180 || x.UncertaintyMeters < 0) {
		return fmt.Errorf("config: invalid simulated position")
	}
	return nil
}

const aperLimit = 1 << 20
