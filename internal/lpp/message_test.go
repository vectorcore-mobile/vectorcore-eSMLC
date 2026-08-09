package lpp

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/vectorcore/esmlc/internal/lpp/capability"
	"github.com/vectorcore/esmlc/internal/lpp/location"
	"github.com/vectorcore/esmlc/internal/uper"
)

type manifest struct {
	Fixtures []struct {
		Name   string `json:"name"`
		Binary string `json:"binary_file"`
		Bits   int    `json:"bit_length"`
		Unused int    `json:"unused_trailing_bits"`
	} `json:"fixtures"`
	Negative []struct {
		Name     string `json:"name"`
		Rejected bool   `json:"rejected"`
	} `json:"negative_cases"`
}

func TestFixtureOracle(t *testing.T) {
	root := filepath.Join("..", "..", "tools", "specs", "lpp", "fixtures", "r16.4.0")
	raw, e := os.ReadFile(filepath.Join(root, "manifest.json"))
	if e != nil {
		t.Fatal(e)
	}
	var m manifest
	if e = json.Unmarshal(raw, &m); e != nil {
		t.Fatal(e)
	}
	if len(m.Fixtures) != 11 || len(m.Negative) != 4 {
		t.Fatal("oracle count")
	}
	for _, n := range m.Negative {
		if !n.Rejected {
			t.Fatal(n.Name)
		}
	}
	for _, f := range m.Fixtures {
		t.Run(f.Name, func(t *testing.T) {
			data, e := os.ReadFile(filepath.Join(root, f.Binary))
			if e != nil {
				t.Fatal(e)
			}
			got, e := DecodeMessage(data, f.Bits)
			if e != nil {
				t.Fatal(e)
			}
			assertMeaning(t, f.Name, got)
			encoded, e := EncodeMessage(got)
			if e != nil {
				t.Fatal(e)
			}
			if string(encoded.Bytes) != string(data) || encoded.BitLength != f.Bits || int(encoded.UnusedBits) != f.Unused {
				t.Fatalf("encoded=%x %d/%d", encoded.Bytes, encoded.BitLength, encoded.UnusedBits)
			}
			again, e := DecodeMessage(encoded.Bytes, encoded.BitLength)
			if e != nil || !equalMessage(got, again) {
				t.Fatalf("round trip %v %+v", e, again)
			}
		})
	}
}
func assertMeaning(t *testing.T, n string, m Message) {
	t.Helper()
	switch n {
	case "minimal-no-transaction":
		if m.TransactionID != nil || m.EndTransaction || m.Body != nil {
			t.Fatal(m)
		}
	case "transaction-location-server-0":
		if m.TransactionID == nil || m.TransactionID.Initiator != InitiatorLocationServer || m.TransactionID.TransactionNumber != 0 || m.EndTransaction {
			t.Fatal(m)
		}
	case "transaction-target-device-255-end":
		if m.TransactionID == nil || m.TransactionID.Initiator != InitiatorTargetDevice || m.TransactionID.TransactionNumber != 255 || !m.EndTransaction {
			t.Fatal(m)
		}
	case "sequence-zero-ack-requested":
		if m.SequenceNumber == nil || *m.SequenceNumber != 0 || m.Acknowledgement == nil || !m.Acknowledgement.Requested {
			t.Fatal(m)
		}
	case "sequence-255-ack-indicator":
		if m.SequenceNumber == nil || *m.SequenceNumber != 255 || m.Acknowledgement == nil || m.Acknowledgement.Requested || m.Acknowledgement.Indicator == nil || *m.Acknowledgement.Indicator != 255 {
			t.Fatal(m)
		}
	default:
		want := map[string]BodyKind{"request-capabilities-r9-empty": 0, "provide-capabilities-r9-empty": 1, "request-location-information-r9-empty": 4, "provide-location-information-r9-empty": 5, "abort-r9-empty": 6, "error-r9-empty": 7}[n]
		if m.Body == nil || m.Body.Kind != want {
			t.Fatal(m)
		}
	}
}
func equalMessage(a, b Message) bool {
	if a.EndTransaction != b.EndTransaction || (a.TransactionID == nil) != (b.TransactionID == nil) || (a.SequenceNumber == nil) != (b.SequenceNumber == nil) || (a.Acknowledgement == nil) != (b.Acknowledgement == nil) || (a.Body == nil) != (b.Body == nil) {
		return false
	}
	if a.TransactionID != nil && *a.TransactionID != *b.TransactionID {
		return false
	}
	if a.SequenceNumber != nil && *a.SequenceNumber != *b.SequenceNumber {
		return false
	}
	if a.Acknowledgement != nil {
		if a.Acknowledgement.Requested != b.Acknowledgement.Requested || (a.Acknowledgement.Indicator == nil) != (b.Acknowledgement.Indicator == nil) {
			return false
		}
		if a.Acknowledgement.Indicator != nil && *a.Acknowledgement.Indicator != *b.Acknowledgement.Indicator {
			return false
		}
	}
	return a.Body == nil || a.Body.Kind == b.Body.Kind
}
func TestMalformed(t *testing.T) {
	if _, e := DecodeMessage([]byte{0}, 9); e == nil {
		t.Fatal("bit length")
	}
	if _, e := DecodeMessage([]byte{0x04}, 5); !errors.Is(e, ErrMalformed) {
		t.Fatal(e)
	}
	if _, e := EncodeMessage(Message{Body: &Body{Kind: 2}}); !errors.Is(e, ErrUnsupportedBody) {
		t.Fatal(e)
	}
	if _, e := DecodeMessage([]byte{0x10}, 5); e == nil {
		t.Fatal("truncation")
	}
}

func TestMessageECIDRequestLocationPayload(t *testing.T) {
	requested, err := uper.NewBitString([]byte{0xe0}, 3)
	if err != nil {
		t.Fatal(err)
	}
	in := Message{Body: &Body{Kind: BodyRequestLocationInformation, RequestLocationInformation: &location.RequestLocationInformationR9IEs{ECID: &location.ECIDRequestLocationInformation{RequestedMeasurements: requested}}}}
	encoded, err := EncodeMessage(in)
	if err != nil {
		t.Fatal(err)
	}
	out, err := DecodeMessage(encoded.Bytes, encoded.BitLength)
	if err != nil {
		t.Fatal(err)
	}
	if out.Body == nil || out.Body.RequestLocationInformation == nil || out.Body.RequestLocationInformation.ECID == nil {
		t.Fatal("missing location payload")
	}
	if !out.Body.RequestLocationInformation.ECID.RequestedMeasurements.Equal(requested) {
		t.Fatal("requested measurements changed")
	}
}

// TestAbortRealUEFixture guards the fix for a real bug found against a real
// UE: a live iPhone declined an LPP RequestCapabilities with an Abort
// message carrying commonIEs.abortCause, and this package's decodeBody
// treated commonIEs being present at all as an unsupported IE ("malformed
// encoding") instead of decoding it. These exact bytes were captured from
// the live SLs association.
func TestAbortRealUEFixture(t *testing.T) {
	data, err := hex.DecodeString("d001003040")
	if err != nil {
		t.Fatal(err)
	}
	out, err := DecodeMessageOctets(data)
	if err != nil {
		t.Fatalf("real UE Abort response must decode, got: %v", err)
	}
	if out.Body == nil || out.Body.Kind != BodyAbort || out.Body.Abort == nil {
		t.Fatalf("unexpected decoded message: %#v", out)
	}
	if out.Body.Abort.Cause == nil || *out.Body.Abort.Cause != AbortCauseUndefined {
		t.Fatalf("cause = %v, want AbortCauseUndefined", out.Body.Abort.Cause)
	}
}

// TestAbortWithCauseRoundTrip guards commonIEs.abortCause round-tripping
// correctly on the encode side too — encodeBody had the same bug, encoding
// zero bits for cause instead of a presence bit plus the enumerated value.
func TestAbortWithCauseRoundTrip(t *testing.T) {
	for _, cause := range []AbortCause{AbortCauseUndefined, AbortCauseStopPeriodicReporting, AbortCauseTargetDeviceAbort, AbortCauseNetworkAbort} {
		c := cause
		in := Message{Body: &Body{Kind: BodyAbort, Abort: &AbortIEs{Cause: &c}}}
		encoded, err := EncodeMessage(in)
		if err != nil {
			t.Fatal(err)
		}
		out, err := DecodeMessage(encoded.Bytes, encoded.BitLength)
		if err != nil {
			t.Fatalf("cause=%d: %v", cause, err)
		}
		if out.Body == nil || out.Body.Kind != BodyAbort || out.Body.Abort == nil || out.Body.Abort.Cause == nil || *out.Body.Abort.Cause != cause {
			t.Fatalf("cause=%d: round-trip mismatch: %#v", cause, out.Body)
		}
	}
}

// TestAbortEmptyFixture is the pre-existing minimum-Abort fixture
// (tools/specs/lpp/fixtures/r16.4.0/abort-r9-empty.uper, commonIEs absent) —
// already covered by TestFixtureOracle, but asserted directly here as an
// explicit regression anchor for the absent-cause case alongside the
// present-cause cases above.
func TestAbortEmptyFixture(t *testing.T) {
	data, err := hex.DecodeString("1980")
	if err != nil {
		t.Fatal(err)
	}
	out, err := DecodeMessage(data, 15)
	if err != nil {
		t.Fatal(err)
	}
	if out.Body == nil || out.Body.Kind != BodyAbort || out.Body.Abort == nil || out.Body.Abort.Cause != nil {
		t.Fatalf("unexpected decoded message: %#v", out)
	}
}

// TestErrorRealUEFixture guards the fix for a real bug found against a real
// UE: a live iPhone responded to a RequestLocationInformation with a
// legitimate Error message carrying commonIEsError.errorCause, and this
// package's decodeBody treated commonIEsError being present at all as an
// unsupported IE ("malformed encoding") instead of decoding it — the same
// bug, in the same shape, as the Abort case above. These exact bytes were
// captured from the live SLs association.
func TestErrorRealUEFixture(t *testing.T) {
	data, err := hex.DecodeString("f003014e50")
	if err != nil {
		t.Fatal(err)
	}
	out, err := DecodeMessageOctets(data)
	if err != nil {
		t.Fatalf("real UE Error response must decode, got: %v", err)
	}
	if out.Body == nil || out.Body.Kind != BodyError || out.Body.Error == nil {
		t.Fatalf("unexpected decoded message: %#v", out)
	}
	if out.Body.Error.Cause == nil || *out.Body.Error.Cause != ErrorCauseIncorrectDataValue {
		t.Fatalf("cause = %v, want ErrorCauseIncorrectDataValue", out.Body.Error.Cause)
	}
}

// TestErrorWithCauseRoundTrip guards commonIEsError.errorCause round-tripping
// correctly on the encode side too — encodeBody had the same bug, encoding
// zero bits for cause instead of a presence bit plus the enumerated value.
func TestErrorWithCauseRoundTrip(t *testing.T) {
	for _, cause := range []ErrorCause{ErrorCauseUndefined, ErrorCauseLPPMessageHeaderError, ErrorCauseLPPMessageBodyError, ErrorCauseEPDUError, ErrorCauseIncorrectDataValue} {
		c := cause
		in := Message{Body: &Body{Kind: BodyError, Error: &ErrorIEs{Cause: &c}}}
		encoded, err := EncodeMessage(in)
		if err != nil {
			t.Fatal(err)
		}
		out, err := DecodeMessage(encoded.Bytes, encoded.BitLength)
		if err != nil {
			t.Fatalf("cause=%d: %v", cause, err)
		}
		if out.Body == nil || out.Body.Kind != BodyError || out.Body.Error == nil || out.Body.Error.Cause == nil || *out.Body.Error.Cause != cause {
			t.Fatalf("cause=%d: round-trip mismatch: %#v", cause, out.Body)
		}
	}
}

// TestErrorEmptyFixture is the pre-existing minimum-Error fixture
// (tools/specs/lpp/fixtures/r16.4.0/error-r9-empty.uper, commonIEsError
// absent) — already covered by TestFixtureOracle, but asserted directly
// here as an explicit regression anchor for the absent-cause case alongside
// the present-cause cases above.
func TestErrorEmptyFixture(t *testing.T) {
	data, err := hex.DecodeString("19c0")
	if err != nil {
		t.Fatal(err)
	}
	out, err := DecodeMessage(data, 13)
	if err != nil {
		t.Fatal(err)
	}
	if out.Body == nil || out.Body.Kind != BodyError || out.Body.Error == nil || out.Body.Error.Cause != nil {
		t.Fatalf("unexpected decoded message: %#v", out)
	}
}

// TestProvideCapabilitiesAGNSSRealUEFixture guards the fix for a real bug
// found against a real UE: an iPhone's A-GNSS ProvideCapabilities response
// legitimately set assistanceDataSupportList, locationCoordinateTypes, and
// velocityTypes (not just gnss-SupportList), and internal/lpp/capability
// previously hard-rejected any of those three as unsupported, reporting a
// genuine capability response as "malformed encoding". These are the exact
// bytes captured from the live SLs association, including the outer LPP
// envelope (TransactionID, EndTransaction, SequenceNumber, Acknowledgement)
// the UE wrapped the ProvideCapabilities body in.
func TestProvideCapabilitiesAGNSSRealUEFixture(t *testing.T) {
	data, err := hex.DecodeString("f0010042087800174027008201b800002080")
	if err != nil {
		t.Fatal(err)
	}
	m, err := DecodeMessageOctets(data)
	if err != nil {
		t.Fatalf("real UE A-GNSS ProvideCapabilities must decode, got: %v", err)
	}
	if !m.EndTransaction {
		t.Fatal("expected EndTransaction")
	}
	if m.Acknowledgement == nil || !m.Acknowledgement.Requested {
		t.Fatalf("expected Acknowledgement.Requested: %#v", m.Acknowledgement)
	}
	if m.Body == nil || m.Body.Kind != BodyProvideCapabilities || m.Body.ProvideCapabilities == nil || m.Body.ProvideCapabilities.AGNSS == nil {
		t.Fatalf("unexpected decoded message: %#v", m.Body)
	}
	a := m.Body.ProvideCapabilities.AGNSS
	if len(a.GNSSSupportList) != 1 || a.GNSSSupportList[0].ID != capability.GNSSIDGPS || !a.GNSSSupportList[0].VelocityMeasurementSupport {
		t.Fatalf("unexpected support list: %#v", a.GNSSSupportList)
	}
	if a.AssistanceData == nil {
		t.Fatal("expected assistanceDataSupportList")
	}
	c := a.AssistanceData.Common
	if c.ReferenceTime == nil || c.ReferenceLocation == nil || c.IonosphericModel == nil || c.EarthOrientation != nil {
		t.Fatalf("unexpected common assistance data support: %#v", c)
	}
	if len(a.AssistanceData.Generic.Elements) != 1 {
		t.Fatalf("unexpected generic assistance data support: %#v", a.AssistanceData.Generic.Elements)
	}
	g := a.AssistanceData.Generic.Elements[0]
	if g.ID != capability.GNSSIDGPS || g.NavigationModelSupport == nil || g.RealTimeIntegritySupport == nil || g.AcquisitionAssistanceSupport == nil || g.AlmanacSupport == nil || g.UTCModelSupport == nil {
		t.Fatalf("unexpected generic assistance data support element: %#v", g)
	}
	if g.SBASID != nil || g.TimeModelsSupport != nil || g.DifferentialCorrectionsSupport != nil || g.DataBitAssistanceSupport != nil || g.AuxiliaryInformationSupport != nil {
		t.Fatalf("unexpected present field on generic assistance data support element: %#v", g)
	}
	if a.LocationCoordinateTypes == nil || !a.LocationCoordinateTypes.EllipsoidPointWithAltitudeAndUncertaintyEllipsoid {
		t.Fatalf("unexpected locationCoordinateTypes: %#v", a.LocationCoordinateTypes)
	}
	if a.VelocityTypes == nil || !a.VelocityTypes.HorizontalWithVerticalVelocityAndUncertainty {
		t.Fatalf("unexpected velocityTypes: %#v", a.VelocityTypes)
	}
}

// TestRequestLocationInformationLocationInformationTypeExtensionBit guards a
// real bug found live: CommonIEsRequestLocationInformation.locationInformationType
// is itself an extensible ENUMERATED (TS 37.355's LocationInformationType has
// a trailing "..."), so it carries its own extension-presence bit distinct
// from the enclosing SEQUENCE's marker — the same shape already fixed for
// AbortCause/ErrorCause. common.go's encode/decode omitted that bit entirely,
// silently shifting every bit of the message after it left by one. A real UE
// (IMSI ...070572) received the corrupted RequestLocationInformation E-SMLC
// sent (decoding gnss-Methods as an all-zero, wrong-length GNSS-ID-Bitmap)
// and correctly rejected it with Error(IncorrectDataValue). The expected
// bytes below were independently produced by asn1tools compiled straight
// from the normative TS 37.355 ASN.1 (tools/specs/lpp/reference-codec), not
// derived from this package.
func TestRequestLocationInformationLocationInformationTypeExtensionBit(t *testing.T) {
	gpsOnly, err := uper.NewBitString([]byte{0x80}, 1)
	if err != nil {
		t.Fatal(err)
	}
	in := Message{
		TransactionID: &TransactionID{Initiator: InitiatorLocationServer, TransactionNumber: 1},
		Body: &Body{
			Kind: BodyRequestLocationInformation,
			RequestLocationInformation: &location.RequestLocationInformationR9IEs{
				Common: &location.CommonRequestLocationInformation{LocationInformationType: location.LocationEstimateRequired},
				AGNSS:  &location.AGNSSRequestLocationInformation{GNSSMethods: gpsOnly},
			},
		},
	}
	encoded, err := EncodeMessage(in)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := hex.EncodeToString(encoded.Bytes), "90022060000080"; got != want {
		t.Fatalf("encode: got %s want %s (asn1tools reference)", got, want)
	}
	out, err := DecodeMessage(encoded.Bytes, encoded.BitLength)
	if err != nil {
		t.Fatal(err)
	}
	if out.Body == nil || out.Body.RequestLocationInformation == nil {
		t.Fatalf("unexpected decoded message: %#v", out)
	}
	got := out.Body.RequestLocationInformation
	if got.Common == nil || got.Common.LocationInformationType != location.LocationEstimateRequired {
		t.Fatalf("unexpected Common: %#v", got.Common)
	}
	if got.AGNSS == nil || got.AGNSS.GNSSMethods.BitLen() != 1 || got.AGNSS.GNSSMethods.Bytes()[0] != 0x80 {
		t.Fatalf("unexpected AGNSS.GNSSMethods: %#v", got.AGNSS)
	}
}

// TestProvideLocationInformationAGNSSRealUEFixture guards a scope gap found
// live: after the LocationInformationType extension-bit fix above let a real
// UE (IMSI ...070572) engage a full A-GNSS exchange, it replied with an
// actual UE-computed GPS fix using the
// ellipsoidPointWithAltitudeAndUncertaintyEllipsoid LocationCoordinates
// shape (root CHOICE index 5) and MeasurementReferenceTime's optional
// gnss-TOD-frac/gnss-TOD-unc fields — none of which this package decoded
// before, so it failed closed as "malformed encoding" even though the
// message was entirely spec-valid. These exact bytes were captured from the
// live SLs association and independently confirmed against asn1tools
// compiled from the normative TS 37.355 ASN.1
// (tools/specs/lpp/reference-codec): latitude 32.6225N, longitude
// -86.295257 (i.e. 86.295257W), altitude 65m height.
func TestProvideLocationInformationAGNSSRealUEFixture(t *testing.T) {
	data, err := hex.DecodeString("f003014a18452e657d42a26e00411e24a86da23220ad2001940080")
	if err != nil {
		t.Fatal(err)
	}
	out, err := DecodeMessageOctets(data)
	if err != nil {
		t.Fatalf("real UE ProvideLocationInformation must decode, got: %v", err)
	}
	if out.Body == nil || out.Body.Kind != BodyProvideLocationInformation || out.Body.ProvideLocationInformation == nil {
		t.Fatalf("unexpected decoded message: %#v", out)
	}
	p := out.Body.ProvideLocationInformation
	if p.Common == nil || p.Common.LocationEstimate == nil {
		t.Fatalf("unexpected Common: %#v", p.Common)
	}
	le := p.Common.LocationEstimate
	if le.Shape != location.ShapePointWithAltitudeAndUncertaintyEllipsoid {
		t.Fatalf("shape = %d, want ShapePointWithAltitudeAndUncertaintyEllipsoid", le.Shape)
	}
	if math.Abs(le.Point.Latitude-32.6225) > 1e-4 || math.Abs(le.Point.Longitude-(-86.295257)) > 1e-4 {
		t.Fatalf("unexpected coordinates: %#v", le.Point)
	}
	if le.AltitudeDirection != location.AltitudeHeight || le.Altitude != 65 {
		t.Fatalf("unexpected altitude: dir=%d alt=%d", le.AltitudeDirection, le.Altitude)
	}
	if le.UncertaintySemiMajor != 15 || le.UncertaintySemiMinor != 9 || le.OrientationMajorAxis != 42 || le.UncertaintyAltitude != 13 || le.Confidence != 90 {
		t.Fatalf("unexpected uncertainty ellipsoid: %#v", le)
	}
	if p.AGNSS == nil || p.AGNSS.GNSSLocationInformation == nil {
		t.Fatalf("unexpected AGNSS: %#v", p.AGNSS)
	}
	m := p.AGNSS.GNSSLocationInformation.MeasurementReferenceTime
	if m.GNSSTODMsec != 1115497 {
		t.Fatalf("unexpected gnss-TOD-msec: %d", m.GNSSTODMsec)
	}
	if m.GNSSTODFrac == nil || *m.GNSSTODFrac != 0 {
		t.Fatalf("unexpected gnss-TOD-frac: %v", m.GNSSTODFrac)
	}
	if m.GNSSTODUnc == nil || *m.GNSSTODUnc != 101 {
		t.Fatalf("unexpected gnss-TOD-unc: %v", m.GNSSTODUnc)
	}

	// Round-trip: re-encoding the decoded value must reproduce the exact
	// real-UE wire bytes.
	encoded, err := EncodeMessage(out)
	if err != nil {
		t.Fatal(err)
	}
	if got := hex.EncodeToString(encoded.Bytes); got != "f003014a18452e657d42a26e00411e24a86da23220ad2001940080" {
		t.Fatalf("round-trip encode mismatch: got %s", got)
	}
}

func TestMessageECIDProvideLocationFixture(t *testing.T) {
	// Independent whole-LPP fixture generated under tools/specs/lpp/analysis;
	// it contains a target-device transaction and all root ECID optionals.
	data, err := hex.DecodeString("920f2809007c0200100848d159c0192aac317ff8")
	if err != nil {
		t.Fatal(err)
	}
	out, err := DecodeMessage(data, 157)
	if err != nil {
		t.Fatal(err)
	}
	if out.TransactionID == nil || out.TransactionID.Initiator != InitiatorTargetDevice || out.TransactionID.TransactionNumber != 7 || !out.EndTransaction || out.Body == nil || out.Body.Kind != BodyProvideLocationInformation || out.Body.ProvideLocationInformation == nil || out.Body.ProvideLocationInformation.ECID == nil {
		t.Fatalf("unexpected decoded message: %#v", out)
	}
	encoded, err := EncodeMessage(out)
	if err != nil {
		t.Fatal(err)
	}
	if encoded.BitLength != 157 || string(encoded.Bytes) != string(data) {
		t.Fatalf("encoded=%x/%d", encoded.Bytes, encoded.BitLength)
	}
}

func TestDecodeMessageOctetsFindsUniqueUPERPadding(t *testing.T) {
	// This independently generated LPP message has 19 meaningful bits in three
	// transport octets. LCS-AP carries the octets, not that bit count.
	data, err := hex.DecodeString("110000")
	if err != nil {
		t.Fatal(err)
	}
	m, err := DecodeMessageOctets(data)
	if err != nil {
		t.Fatal(err)
	}
	if m.Body == nil || m.Body.Kind != BodyRequestLocationInformation {
		t.Fatalf("%#v", m)
	}
}
