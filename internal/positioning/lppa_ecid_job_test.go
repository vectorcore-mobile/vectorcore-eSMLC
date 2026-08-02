package positioning

import (
	"testing"
	"time"

	"github.com/vectorcore/esmlc/internal/lppa"
)

func lppaPolicy(t *testing.T, enabled bool) Policy {
	t.Helper()
	return Policy{LPPaECID: LPPaECIDPolicy{Enabled: enabled}}
}

// TestJobStartsLPPaECIDWithoutUERoundTrip proves LPPaECID bypasses both the
// LPP capability exchange and the LPPSupported gate: it must start even when
// the LCS-AP request reports the UE does not support LPP at all.
func TestJobStartsLPPaECIDWithoutUERoundTrip(t *testing.T) {
	now := time.Unix(0, 0)
	m := New(lppaPolicy(t, true))
	request := req()
	unsupported := false
	request.LPPSupported = &unsupported
	r, err := m.Start(request, proc(t), now)
	if err != nil || r.Snapshot.Method != MethodLPPaECID || r.Snapshot.State != AwaitingLocationInformation {
		t.Fatalf("start %#v %v", r, err)
	}
	if r.LPPa == nil || r.LPPa.Kind != LPPaSendInitiationRequest || r.LPPa.ESMLCMeasurementID != 1 {
		t.Fatalf("expected initiation request action: %#v", r.LPPa)
	}
	if len(r.Actions) != 0 {
		t.Fatalf("expected no LPP actions, got %#v", r.Actions)
	}
}

func TestJobDeliversLPPaAccessPointPositionThroughResponse(t *testing.T) {
	now := time.Unix(0, 0)
	m := NewWithEstimator(lppaPolicy(t, true), CombinedEstimator{LPPaECID: LPPaECIDEstimator{}})
	start, err := m.Start(req(), proc(t), now)
	if err != nil {
		t.Fatal(err)
	}
	result := lppa.ECIDMeasurementResult{
		ServingCellID:       lppa.ECGI{PLMNIdentity: [3]byte{0x00, 0xf1, 0x10}, CellIdentity: 1},
		ServingCellTAC:      [2]byte{0x10, 0x01},
		AccessPointPosition: &lppa.AccessPointPosition{Latitude: 38, Longitude: -90, Confidence: 68, UncertaintySemiMajor: 10},
	}
	resp := lppa.InitiationResponse{ESMLCMeasurementID: 1, ENBMeasurementID: 9, Result: &result}
	final, err := m.ApplyLPPaInitiationResponse(req().Scope, start.LPPa.TransactionID, resp, now)
	if err != nil {
		t.Fatal(err)
	}
	if final.Snapshot.State != EstimateAvailable || final.Snapshot.Final == nil || final.Snapshot.Final.Kind != FinalEstimateAvailable {
		t.Fatalf("expected delivered estimate, got %#v", final.Snapshot.Final)
	}
	if final.Snapshot.Final.Estimate.Source != EstimateSourceLPPaAccessPointPosition || final.Snapshot.Final.Estimate.Latitude != 38 || final.Snapshot.Final.Estimate.Longitude != -90 {
		t.Fatalf("wrong estimate: %#v", final.Snapshot.Final.Estimate)
	}
	if final.LPPa == nil || final.LPPa.Kind != LPPaSendTerminationCommand || final.LPPa.ENBMeasurementID != 9 {
		t.Fatalf("expected termination command action, got %#v", final.LPPa)
	}
}

// TestJobAwaitsUnsolicitedLPPaReportWhenResponseHasNoResult exercises the
// path where the Initiation Response merely acknowledges the request and the
// actual measurement arrives later via an unsolicited Report.
func TestJobAwaitsUnsolicitedLPPaReportWhenResponseHasNoResult(t *testing.T) {
	now := time.Unix(0, 0)
	catalog := NewCatalogStore("", 0, func() time.Time { return now })
	m := NewWithEstimator(lppaPolicy(t, true), CombinedEstimator{LPPaECID: LPPaECIDEstimator{Store: catalog}})
	request := req()
	request.ServingECGI = [7]byte{0, 0xf1, 0x10, 0, 0, 0, 1}
	start, err := m.Start(request, proc(t), now)
	if err != nil {
		t.Fatal(err)
	}
	resp := lppa.InitiationResponse{ESMLCMeasurementID: 1, ENBMeasurementID: 9}
	ack, err := m.ApplyLPPaInitiationResponse(request.Scope, start.LPPa.TransactionID, resp, now)
	if err != nil {
		t.Fatal(err)
	}
	if ack.Snapshot.State != AwaitingLocationInformation || ack.Snapshot.Final != nil {
		t.Fatalf("expected job to remain active awaiting a report: %#v", ack.Snapshot)
	}
	report := lppa.Report{ESMLCMeasurementID: 1, ENBMeasurementID: 9, Result: lppa.ECIDMeasurementResult{
		ServingCellID: lppa.ECGI{PLMNIdentity: [3]byte{0, 0xf1, 0x10}, CellIdentity: 1}, ServingCellTAC: [2]byte{0, 1},
	}}
	final, err := m.ApplyLPPaReport(request.Scope, report, now)
	if err != nil {
		t.Fatal(err)
	}
	// No catalog entry was loaded (empty file), so this must fail closed
	// rather than fabricate a position.
	if final.Snapshot.State != EstimationUnavailable {
		t.Fatalf("expected estimation-unavailable without a catalog entry: %#v", final.Snapshot)
	}
}

func TestJobLPPaInitiationFailureNeedsNoTermination(t *testing.T) {
	now := time.Unix(0, 0)
	m := New(lppaPolicy(t, true))
	start, err := m.Start(req(), proc(t), now)
	if err != nil {
		t.Fatal(err)
	}
	final, err := m.ApplyLPPaInitiationFailure(req().Scope, start.LPPa.TransactionID, lppa.InitiationFailure{ESMLCMeasurementID: 1, Cause: lppa.Cause{Branch: lppa.CauseMisc}}, now)
	if err != nil {
		t.Fatal(err)
	}
	if final.Snapshot.State != ProcedureFailed || final.LPPa != nil {
		t.Fatalf("expected procedure failure with no termination action: %#v %#v", final.Snapshot, final.LPPa)
	}
}

func TestJobLPPaFailureIndicationEndsMeasurementWithoutTermination(t *testing.T) {
	now := time.Unix(0, 0)
	m := New(lppaPolicy(t, true))
	start, err := m.Start(req(), proc(t), now)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := m.ApplyLPPaInitiationResponse(req().Scope, start.LPPa.TransactionID, lppa.InitiationResponse{ESMLCMeasurementID: 1, ENBMeasurementID: 9}, now); err != nil {
		t.Fatal(err)
	}
	final, err := m.ApplyLPPaFailureIndication(req().Scope, lppa.FailureIndication{ESMLCMeasurementID: 1, ENBMeasurementID: 9, Cause: lppa.Cause{Branch: lppa.CauseMisc}}, now)
	if err != nil {
		t.Fatal(err)
	}
	if final.Snapshot.State != ProcedureFailed || final.LPPa != nil {
		t.Fatalf("expected no termination command after an eNB-reported failure: %#v %#v", final.Snapshot, final.LPPa)
	}
}

func TestLPPaExistingDeploymentsUnaffectedWhenDisabled(t *testing.T) {
	now := time.Unix(0, 0)
	m := New(policy(t, true))
	r, err := m.Start(req(), proc(t), now)
	if err != nil || r.Snapshot.Method != MethodECID || r.Snapshot.State != AwaitingCapabilities || r.LPPa != nil {
		t.Fatalf("LPPaECID disabled must not change ECID-only behaviour: %#v %v", r, err)
	}
}
