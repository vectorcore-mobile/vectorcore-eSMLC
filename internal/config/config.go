package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"net"
	"os"
	"time"
)

type Config struct {
	Service       Service       `yaml:"service"`
	SLs           SLs           `yaml:"sls"`
	Positioning   Positioning   `yaml:"positioning"`
	Observability Observability `yaml:"observability"`
}
type Service struct {
	Name     string `yaml:"name"`
	LogLevel string `yaml:"log_level"`
	// LogFile is where all logging is written, regardless of the -d console
	// debug flag; console output is a separate, additive concern (see
	// internal/logging).
	LogFile string `yaml:"log_file"`
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
	// PruneInterval bounds how often positioning.Manager.Prune runs. Without
	// this sweep, a job whose UE/eNB goes silent (no further LPP/LPPa event
	// ever arrives for its correlation) is only ever expired reactively, on
	// the next event for that same correlation — one that, by definition,
	// never comes. Left unswept, such jobs (and the LPP session/transaction
	// state keyed alongside them) accumulate for the life of the process.
	PruneInterval time.Duration `yaml:"prune_interval"`
}

// Observability is a small HTTP listener separate from the SCTP transport,
// exposing Prometheus-format metrics, health/readiness endpoints, and a
// POST /admin/reload-catalog action. It is disabled by default: existing
// deployments are unaffected until an operator opts in. The default listen
// address is loopback-only, not 0.0.0.0: unlike a read-only metrics
// endpoint, this listener also carries a mutating admin action with no
// authentication of its own, so binding wider than localhost is an
// explicit operator choice, not the default.
type Observability struct {
	Enabled       bool   `yaml:"enabled"`
	ListenAddress string `yaml:"listen_address"`
	Port          int    `yaml:"port"`
}

type Positioning struct {
	Method     string         `yaml:"method"`
	ECID       ECIDPolicy     `yaml:"ecid"`
	OTDOA      OTDOAPolicy    `yaml:"otdoa"`
	AGNSS      AGNSSPolicy    `yaml:"agnss"`
	LPPaECID   LPPaECIDPolicy `yaml:"lppa_ecid"`
	Simulation Simulation     `yaml:"simulation"`
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

// OTDOAPolicy is local operator authorization only, mirroring ECIDPolicy's
// intent. It has no requested-measurements field: OTDOA's bounded
// RequestLocationInformation form carries only assistanceAvailability, which
// this service always sends false (no assistance-data source exists). The
// multilateration estimator reuses ECIDPolicy.CellDataFile/CellDataMaxAge —
// the same operator-maintained cell catalog gives ECID a serving-cell
// reference point and gives OTDOA the reference/neighbour cell positions its
// solver needs, so a second catalog file is not introduced here.
type OTDOAPolicy struct {
	Enabled bool `yaml:"enabled"`
}

// AGNSSPolicy is local operator authorization only, mirroring
// ECIDPolicy/OTDOAPolicy's intent. This service only ever requests GPS in
// UE-based (MS-based) mode: the UE reports its own already-computed
// position, so there is no requested-measurements field and no cell catalog
// dependency (unlike ECID/OTDOA, the estimator here does no computation of
// its own). There is no assistance-data source, so a real UE may be unable
// to produce a position at all without one — see docs/limitations.md.
type AGNSSPolicy struct {
	Enabled bool `yaml:"enabled"`
}

// LPPaECIDPolicy is local operator authorization only, mirroring
// ECIDPolicy/OTDOAPolicy/AGNSSPolicy's intent. Unlike those methods this one
// needs no UE round trip at all: the E-SMLC asks the eNB directly, over
// LPPa, for the serving cell's identity and (if the eNB has it) its own
// known antenna position. When enabled, it takes priority over every
// UE-based method (ECID/OTDOA/A-GNSS) since it is faster and does not depend
// on UE LPP support; it reuses ECIDPolicy.CellDataFile/CellDataMaxAge as a
// fallback when the eNB does not report its own antenna position.
type LPPaECIDPolicy struct {
	Enabled bool `yaml:"enabled"`
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
	return Config{
		Service:       Service{Name: "vectorcore-esmlc", LogLevel: "info", LogFile: "esmlc.log"},
		SLs:           SLs{Enabled: true, ListenAddress: "0.0.0.0", Port: 9082, ExpectedPPID: 29, MaxAssociations: 32, MaxSessions: 10000, SessionTimeout: 10 * time.Second, MaxMessageSize: 1 << 20, PruneInterval: 30 * time.Second},
		Positioning:   Positioning{Method: "gnss", Simulation: Simulation{Enabled: false}},
		Observability: Observability{Enabled: false, ListenAddress: "127.0.0.1", Port: 9090},
	}
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
	if c.Service.LogFile == "" {
		return fmt.Errorf("config: service.log_file is required")
	}
	// Validated regardless of sls.enabled: observability is an independent
	// HTTP listener, not part of the SCTP transport.
	if c.Observability.Enabled {
		o := c.Observability
		if net.ParseIP(o.ListenAddress) == nil {
			return fmt.Errorf("config: observability.listen_address must be an IP address")
		}
		if o.Port < 1 || o.Port > 65535 {
			return fmt.Errorf("config: observability.port must be 1..65535")
		}
		if c.SLs.Enabled && o.Port == c.SLs.Port {
			return fmt.Errorf("config: observability.port must differ from sls.port")
		}
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
	if s.PruneInterval <= 0 {
		return fmt.Errorf("config: prune interval must be positive")
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
