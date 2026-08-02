package positioning

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// ServingCellRecord is operator-supplied authoritative geometry for one E-CGI.
// CoverageUncertainty is the standards GAD uncertainty code selected by the
// operator to conservatively bound the serving-cell reference estimate.
type ServingCellRecord struct {
	ECGI                [7]byte
	Latitude            float64
	Longitude           float64
	CoverageUncertainty uint8
	Source              string
	UpdatedAt           time.Time
}

type CellCatalog struct {
	version       string
	cells         map[[7]byte]ServingCellRecord
	maxAge        time.Duration
	loadedAt      time.Time
	oldestUpdated time.Time
	newestUpdated time.Time
}

type catalogFile struct {
	SchemaVersion uint8  `yaml:"schema_version"`
	Version       string `yaml:"version"`
	Cells         []struct {
		ECGI                string  `yaml:"ecgi"`
		Latitude            float64 `yaml:"latitude"`
		Longitude           float64 `yaml:"longitude"`
		CoverageUncertainty uint8   `yaml:"coverage_uncertainty_code"`
		Source              string  `yaml:"source"`
		UpdatedAt           string  `yaml:"updated_at"`
	} `yaml:"cells"`
}

func LoadCellCatalog(path string, maxAge time.Duration, now time.Time) (*CellCatalog, error) {
	if path == "" || maxAge <= 0 {
		return nil, fmt.Errorf("positioning: cell catalog path and positive maximum age are required")
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("positioning: read cell catalog: %w", err)
	}
	var in catalogFile
	decoder := yaml.NewDecoder(bytes.NewReader(b))
	decoder.KnownFields(true)
	if err := decoder.Decode(&in); err != nil {
		return nil, fmt.Errorf("positioning: parse cell catalog: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return nil, fmt.Errorf("positioning: catalog must contain one document")
	}
	if in.SchemaVersion != 1 || in.Version == "" || len(in.Cells) == 0 {
		return nil, fmt.Errorf("positioning: supported catalog schema version, version, and cells are required")
	}
	catalog := &CellCatalog{version: in.Version, cells: make(map[[7]byte]ServingCellRecord, len(in.Cells)), maxAge: maxAge, loadedAt: now}
	for i, cell := range in.Cells {
		ecgi, err := parseECGI(cell.ECGI)
		if err != nil {
			return nil, fmt.Errorf("positioning: cell %d: %w", i, err)
		}
		updated, err := time.Parse(time.RFC3339, cell.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("positioning: cell %d updated_at: %w", i, err)
		}
		record := ServingCellRecord{ECGI: ecgi, Latitude: cell.Latitude, Longitude: cell.Longitude, CoverageUncertainty: cell.CoverageUncertainty, Source: cell.Source, UpdatedAt: updated}
		if err := validateServingCell(record, now, maxAge); err != nil {
			return nil, fmt.Errorf("positioning: cell %d: %w", i, err)
		}
		if _, exists := catalog.cells[ecgi]; exists {
			return nil, fmt.Errorf("positioning: duplicate ECGI %x", ecgi)
		}
		catalog.cells[ecgi] = record
		if catalog.oldestUpdated.IsZero() || updated.Before(catalog.oldestUpdated) {
			catalog.oldestUpdated = updated
		}
		if catalog.newestUpdated.IsZero() || updated.After(catalog.newestUpdated) {
			catalog.newestUpdated = updated
		}
	}
	return catalog, nil
}

func (c *CellCatalog) recordCount() int {
	if c == nil {
		return 0
	}
	return len(c.cells)
}

func (c *CellCatalog) Version() string {
	if c == nil {
		return ""
	}
	return c.version
}
func (c *CellCatalog) Lookup(ecgi [7]byte, now time.Time) (ServingCellRecord, bool) {
	v, ok, _ := c.lookup(ecgi, now)
	return v, ok
}

func (c *CellCatalog) lookup(ecgi [7]byte, now time.Time) (ServingCellRecord, bool, bool) {
	if c == nil {
		return ServingCellRecord{}, false, false
	}
	v, ok := c.cells[ecgi]
	if !ok {
		return ServingCellRecord{}, false, false
	}
	if now.Sub(v.UpdatedAt) > c.maxAge {
		return ServingCellRecord{}, false, true
	}
	return v, true, false
}
func parseECGI(v string) ([7]byte, error) {
	var out [7]byte
	b, err := hex.DecodeString(v)
	if err != nil || len(b) != len(out) {
		return out, fmt.Errorf("ECGI must be fourteen hexadecimal characters")
	}
	copy(out[:], b)
	return out, nil
}
func validateServingCell(v ServingCellRecord, now time.Time, maxAge time.Duration) error {
	if err := (GeographicEstimate{Latitude: v.Latitude, Longitude: v.Longitude, HorizontalUncertainty: v.CoverageUncertainty, Source: EstimateSourceAuthoritativeServingCell, Timestamp: v.UpdatedAt}).Validate(); err != nil {
		return err
	}
	if v.Source == "" || v.UpdatedAt.After(now) || now.Sub(v.UpdatedAt) > maxAge {
		return fmt.Errorf("cell source or freshness is invalid")
	}
	return nil
}

// ServingCellEstimator returns a deliberately coarse, authoritative cell
// reference estimate. It does not infer UE distance, bearing, or altitude.
type ServingCellEstimator struct{ Catalog *CellCatalog }

func (e ServingCellEstimator) Estimate(request Request, method MethodResult, now time.Time) EstimationResult {
	if method.Method != MethodECID || method.ECID == nil {
		return EstimationResult{Failure: InsufficientNetworkData}
	}
	cell, ok := e.Catalog.Lookup(request.ServingECGI, now)
	if !ok {
		return EstimationResult{Failure: InsufficientNetworkData}
	}
	estimate := GeographicEstimate{Latitude: cell.Latitude, Longitude: cell.Longitude, HorizontalUncertainty: cell.CoverageUncertainty, Source: EstimateSourceAuthoritativeServingCell, Timestamp: now}
	return EstimationResult{Estimate: &estimate}
}

// CatalogStatus exposes bounded, non-sensitive operator state. It never
// includes cell coordinates or UE-specific data.
type CatalogStatus struct {
	Configured             bool
	Source                 string
	ActiveVersion          string
	RecordCount            int
	LoadedAt               time.Time
	OldestUpdatedAt        time.Time
	NewestUpdatedAt        time.Time
	ReloadSuccesses        uint64
	ReloadFailures         uint64
	LastReloadAt           time.Time
	LastReloadError        string
	AuthoritativeEstimates uint64
	MissingCell            uint64
	StaleData              uint64
}

// CatalogReloadResult describes one operator-triggered candidate publication.
type CatalogReloadResult struct {
	ActiveChanged    bool
	ActiveVersion    string
	CandidateVersion string
	RecordCount      int
	Error            string
}

// CatalogStore serializes reloads and publishes immutable catalogs. Estimator
// calls take a single snapshot at invocation, so a reload cannot mix records
// from two catalog versions within one estimate.
type CatalogStore struct {
	source string
	maxAge time.Duration
	now    func() time.Time

	mu       sync.RWMutex
	reloadMu sync.Mutex
	active   *CellCatalog
	status   CatalogStatus
}

func NewCatalogStore(source string, maxAge time.Duration, now func() time.Time) *CatalogStore {
	if now == nil {
		now = time.Now
	}
	return &CatalogStore{source: source, maxAge: maxAge, now: now, status: CatalogStatus{Configured: source != "", Source: source}}
}

func (s *CatalogStore) Reload() CatalogReloadResult {
	if s == nil || s.source == "" || s.maxAge <= 0 {
		return CatalogReloadResult{Error: "catalog is not configured"}
	}
	s.reloadMu.Lock()
	defer s.reloadMu.Unlock()
	now := s.now()
	candidate, err := LoadCellCatalog(s.source, s.maxAge, now)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.LastReloadAt = now
	if err != nil {
		s.status.ReloadFailures++
		s.status.LastReloadError = "catalog validation failed"
		return CatalogReloadResult{ActiveVersion: s.status.ActiveVersion, Error: s.status.LastReloadError}
	}
	s.active = candidate
	s.status.ActiveVersion = candidate.Version()
	s.status.RecordCount = candidate.recordCount()
	s.status.LoadedAt = candidate.loadedAt
	s.status.OldestUpdatedAt = candidate.oldestUpdated
	s.status.NewestUpdatedAt = candidate.newestUpdated
	s.status.ReloadSuccesses++
	s.status.LastReloadError = ""
	return CatalogReloadResult{ActiveChanged: true, ActiveVersion: candidate.Version(), CandidateVersion: candidate.Version(), RecordCount: candidate.recordCount()}
}

func (s *CatalogStore) snapshot() *CellCatalog {
	if s == nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.active
}

func (s *CatalogStore) Status() CatalogStatus {
	if s == nil {
		return CatalogStatus{}
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.status
}

// ServingCellCatalogEstimator reads exactly one catalog snapshot for each
// invocation. A successful reload is therefore visible only to later work.
type ServingCellCatalogEstimator struct{ Store *CatalogStore }

func (e ServingCellCatalogEstimator) Estimate(request Request, method MethodResult, now time.Time) EstimationResult {
	if method.Method != MethodECID || method.ECID == nil {
		return EstimationResult{Failure: InsufficientNetworkData}
	}
	catalog := e.Store.snapshot()
	if catalog == nil {
		return EstimationResult{Failure: InsufficientNetworkData}
	}
	cell, ok, stale := catalog.lookup(request.ServingECGI, now)
	if !ok {
		if stale {
			e.Store.recordStaleData()
		} else {
			e.Store.recordMissingCell()
		}
		return EstimationResult{Failure: InsufficientNetworkData}
	}
	e.Store.recordAuthoritativeEstimate()
	estimate := GeographicEstimate{Latitude: cell.Latitude, Longitude: cell.Longitude, HorizontalUncertainty: cell.CoverageUncertainty, Source: EstimateSourceAuthoritativeServingCell, Timestamp: now, CatalogVersion: catalog.Version(), DataSource: cell.Source, RecordUpdatedAt: cell.UpdatedAt}
	return EstimationResult{Estimate: &estimate}
}

func (s *CatalogStore) recordAuthoritativeEstimate() {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.AuthoritativeEstimates++
}
func (s *CatalogStore) recordMissingCell() {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.MissingCell++
}
func (s *CatalogStore) recordStaleData() {
	if s == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status.StaleData++
}
