package positioning

import (
	"encoding/hex"
	"fmt"
	"os"
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
	version string
	cells   map[[7]byte]ServingCellRecord
	maxAge  time.Duration
}

type catalogFile struct {
	Version string `yaml:"version"`
	Cells   []struct {
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
	if err := yaml.Unmarshal(b, &in); err != nil {
		return nil, fmt.Errorf("positioning: parse cell catalog: %w", err)
	}
	if in.Version == "" || len(in.Cells) == 0 {
		return nil, fmt.Errorf("positioning: catalog version and cells are required")
	}
	catalog := &CellCatalog{version: in.Version, cells: make(map[[7]byte]ServingCellRecord, len(in.Cells)), maxAge: maxAge}
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
	}
	return catalog, nil
}

func (c *CellCatalog) Version() string {
	if c == nil {
		return ""
	}
	return c.version
}
func (c *CellCatalog) Lookup(ecgi [7]byte, now time.Time) (ServingCellRecord, bool) {
	if c == nil {
		return ServingCellRecord{}, false
	}
	v, ok := c.cells[ecgi]
	if !ok || now.Sub(v.UpdatedAt) > c.maxAge {
		return ServingCellRecord{}, false
	}
	return v, true
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
