package positioning

import (
	"math"
	"time"

	"github.com/vectorcore/esmlc/internal/lpp/location/result"
)

// speedOfLight is c in metres/second, used to convert an RSTD time
// difference into a range difference.
const speedOfLight = 299792458.0

// earthRadiusMeters is the mean Earth radius used for the local
// equirectangular projection below. It is an approximation appropriate at
// cell-tower scale (a few to tens of kilometres), not for geodesic survey
// work; the same order of simplification the rest of this package already
// makes (see LocationEstimate's linear degree scaling).
const earthRadiusMeters = 6371000.0

// OTDOAEstimator solves 2D hyperbolic multilateration from RSTD
// measurements plus known reference/neighbour cell positions taken from the
// same operator-maintained catalog the ECID authoritative estimator uses.
// It requires every cell it uses (reference and each neighbour) to be
// identified by ECGI in the wire measurement and present in the catalog;
// physCellId alone is not resolved to a specific cell, since PCI is reused
// across a network and resolving it correctly would need the OTDOA
// assistance-data neighbour list this implementation does not have.
//
// This estimator assumes eNB transmission timing is synchronized across the
// network (a standard, explicit OTDOA deployment mode, not a physical
// guarantee): TS 37.355's own RSTD field description defers RSTD's exact
// mapping to TS 36.133 clause 9.1.10.3, which characterizes the measured
// quantity, not any inter-eNB timing offset. Any real inter-cell timing
// offset would come from OTDOA assistance data (expectedRSTD and related
// fields in OTDOA-ReferenceCellInfo/OTDOA-NeighbourCellInfoList), which is
// out of scope for this phase (see docs/roadmap.md Phase 2).
type OTDOAEstimator struct{ Store *CatalogStore }

func (e OTDOAEstimator) Estimate(request Request, method MethodResult, now time.Time) EstimationResult {
	if method.Method != MethodOTDOA || method.OTDOA == nil {
		return EstimationResult{Failure: InsufficientNetworkData}
	}
	catalog := e.Store.snapshot()
	if catalog == nil {
		return EstimationResult{Failure: InsufficientNetworkData}
	}
	signal := method.OTDOA.Signal
	refECGI, ok := signal.CellGlobalIDRef()
	if !ok {
		return EstimationResult{Failure: InsufficientNetworkData}
	}
	refRecord, ok := catalog.Lookup(catalogECGIKey(refECGI), now)
	if !ok {
		return EstimationResult{Failure: InsufficientNetworkData}
	}
	neighbours := signal.NeighbourMeasurements()
	var stations []station
	var rangeDiffs []float64
	for _, nb := range neighbours {
		ecgi, ok := nb.CellGlobalID()
		if !ok {
			continue
		}
		record, ok := catalog.Lookup(catalogECGIKey(ecgi), now)
		if !ok {
			continue
		}
		x, y := projectToLocalMeters(refRecord.Latitude, refRecord.Longitude, record.Latitude, record.Longitude)
		stations = append(stations, station{x: x, y: y})
		rangeDiffs = append(rangeDiffs, nb.RSTD().Duration().Seconds()*speedOfLight)
	}
	if len(stations) < 2 {
		return EstimationResult{Failure: InsufficientNetworkData}
	}
	x, y, rmsResidual, ok := solveHyperbolicTDOA(station{0, 0}, stations, rangeDiffs)
	if !ok {
		return EstimationResult{Failure: InsufficientNetworkData}
	}
	lat, lon := unprojectFromLocalMeters(refRecord.Latitude, refRecord.Longitude, x, y)
	estimate := GeographicEstimate{
		Latitude:              lat,
		Longitude:             lon,
		HorizontalUncertainty: UncertaintyCodeFromMeters(math.Max(rmsResidual, 50)),
		Source:                EstimateSourceOTDOAMultilateration,
		Timestamp:             now,
		CatalogVersion:        catalog.Version(),
	}
	if err := estimate.Validate(); err != nil {
		return EstimationResult{Failure: InsufficientNetworkData}
	}
	return EstimationResult{Estimate: &estimate}
}

// catalogECGIKey converts LPP's decomposed ECGI (MCC digits, MNC digits,
// 28-bit cell identity) into the raw 7-octet PLMN-BCD + cell-identity wire
// form the operator catalog is keyed by (the same form LCS-AP's E-CGI IE
// uses, TS 29.171 clause 9.2's octet layout: MCC digit 2|digit 1, MNC digit
// 3|MCC digit 3, MNC digit 2|MNC digit 1, then 28-bit cell identity plus 4
// zero spare bits). uper.BitString.Bytes() already guarantees the 28-bit
// cell identity's trailing 4 padding bits are zero, so it is used directly.
func catalogECGIKey(v result.ECGI) [7]byte {
	var out [7]byte
	mcc := v.MCC().Digits()
	mnc := v.MNC().Digits()
	mncDigit3 := uint8(0xF)
	if v.MNC().Length() == 3 {
		mncDigit3 = uint8(mnc[2])
	}
	out[0] = uint8(mcc[1])<<4 | uint8(mcc[0])
	out[1] = mncDigit3<<4 | uint8(mcc[2])
	out[2] = uint8(mnc[1])<<4 | uint8(mnc[0])
	copy(out[3:], v.CellIdentity().Bytes())
	return out
}

type station struct{ x, y float64 }

// solveHyperbolicTDOA recovers the 2D position minimizing the sum of squared
// residuals between modeled and measured range differences relative to the
// reference station, via Gauss-Newton least squares. It requires at least
// two non-reference stations (a determined 2D system: each contributes one
// hyperbolic equation in two unknowns) and fails closed on non-convergence
// or degenerate (near-collinear) geometry rather than returning an
// arbitrary point.
func solveHyperbolicTDOA(ref station, neighbours []station, rangeDiff []float64) (x, y, rmsResidual float64, ok bool) {
	if len(neighbours) < 2 || len(neighbours) != len(rangeDiff) {
		return 0, 0, 0, false
	}
	x, y = ref.x, ref.y
	for _, s := range neighbours {
		x += s.x
		y += s.y
	}
	n := float64(len(neighbours) + 1)
	x /= n
	y /= n
	const maxIterations = 50
	const toleranceMetres = 1e-3
	converged := false
	for iter := 0; iter < maxIterations; iter++ {
		d0 := math.Hypot(x-ref.x, y-ref.y)
		if d0 < 1e-6 {
			d0 = 1e-6
		}
		var jtjXX, jtjXY, jtjYY, jtrX, jtrY float64
		for i, s := range neighbours {
			di := math.Hypot(x-s.x, y-s.y)
			if di < 1e-6 {
				di = 1e-6
			}
			resid := (di - d0) - rangeDiff[i]
			jx := (x-s.x)/di - (x-ref.x)/d0
			jy := (y-s.y)/di - (y-ref.y)/d0
			jtjXX += jx * jx
			jtjXY += jx * jy
			jtjYY += jy * jy
			jtrX += jx * resid
			jtrY += jy * resid
		}
		det := jtjXX*jtjYY - jtjXY*jtjXY
		if math.Abs(det) < 1e-9 {
			return 0, 0, 0, false
		}
		dx := (jtjYY*jtrX - jtjXY*jtrY) / det
		dy := (jtjXX*jtrY - jtjXY*jtrX) / det
		x -= dx
		y -= dy
		if math.Hypot(dx, dy) < toleranceMetres {
			converged = true
			break
		}
	}
	if !converged {
		return 0, 0, 0, false
	}
	d0 := math.Hypot(x-ref.x, y-ref.y)
	var sumSq float64
	for i, s := range neighbours {
		di := math.Hypot(x-s.x, y-s.y)
		r := (di - d0) - rangeDiff[i]
		sumSq += r * r
	}
	return x, y, math.Sqrt(sumSq / float64(len(neighbours))), true
}

// projectToLocalMeters and unprojectFromLocalMeters form a local
// equirectangular tangent-plane projection centred on (refLat,refLon),
// accurate enough at cell-tower scale (see earthRadiusMeters doc comment).
func projectToLocalMeters(refLat, refLon, lat, lon float64) (x, y float64) {
	latRad := refLat * math.Pi / 180
	x = (lon - refLon) * math.Pi / 180 * earthRadiusMeters * math.Cos(latRad)
	y = (lat - refLat) * math.Pi / 180 * earthRadiusMeters
	return x, y
}

func unprojectFromLocalMeters(refLat, refLon, x, y float64) (lat, lon float64) {
	latRad := refLat * math.Pi / 180
	lat = refLat + (y/earthRadiusMeters)*180/math.Pi
	lon = refLon + (x/(earthRadiusMeters*math.Cos(latRad)))*180/math.Pi
	return lat, lon
}
