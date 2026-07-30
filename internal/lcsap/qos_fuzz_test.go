package lcsap

import "testing"

func FuzzDecodeQoS(f *testing.F) {
	f.Add([]byte{0})
	f.Add([]byte{0x90})
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = DecodeQoS(data)
	})
}

func FuzzDecodeUEPositioningCapability(f *testing.F) {
	f.Add([]byte{0x40})
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = DecodeUEPositioningCapability(data)
	})
}
