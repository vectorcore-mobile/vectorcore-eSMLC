package positioning

import (
	"testing"
	"time"

	"github.com/vectorcore/esmlc/internal/lpp/location"
	"github.com/vectorcore/esmlc/internal/lpp/procedure"
	"github.com/vectorcore/esmlc/internal/lpp/transaction"
	"github.com/vectorcore/esmlc/internal/uper"
)

func FuzzJobStartPolicy(f *testing.F) {
	f.Add(true, uint8(0x80))
	f.Add(false, uint8(0x80))
	f.Add(true, uint8(0))
	f.Fuzz(func(t *testing.T, enabled bool, bitmap uint8) {
		store, err := transaction.NewStore(transaction.DefaultLimits())
		if err != nil {
			t.Fatal(err)
		}
		p, err := procedure.New(store, procedure.DefaultConfig())
		if err != nil {
			t.Fatal(err)
		}
		request := location.ECIDRequestLocationInformation{}
		bitmap &= 0xe0
		if bitmap != 0 {
			bits, err := uper.NewBitString([]byte{bitmap}, 3)
			if err != nil {
				t.Fatal(err)
			}
			request.RequestedMeasurements = bits
		}
		m := New(Policy{ECID: ECIDPolicy{Enabled: enabled, RequestedMeasurements: request}})
		_, _ = m.Start(Request{Scope: Scope{Association: "fuzz", Correlation: [4]byte{1}}, Deadline: time.Unix(1, 0)}, p, time.Unix(0, 0))
	})
}
