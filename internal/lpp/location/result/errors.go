package result

import "errors"

var (
	ErrPhysicalCellIDOutOfRange          = errors.New("lpp location result: physical cell ID out of range")
	ErrEUTRAARFCNOutOfRange              = errors.New("lpp location result: root E-UTRA ARFCN out of range")
	ErrPhysicalCellIDEncode              = errors.New("lpp location result: physical cell ID encode failed")
	ErrPhysicalCellIDDecode              = errors.New("lpp location result: physical cell ID decode failed")
	ErrEUTRAARFCNEncode                  = errors.New("lpp location result: root E-UTRA ARFCN encode failed")
	ErrEUTRAARFCNDecode                  = errors.New("lpp location result: root E-UTRA ARFCN decode failed")
	ErrRSRPResultOutOfRange              = errors.New("lpp location result: RSRP result out of range")
	ErrRSRQResultOutOfRange              = errors.New("lpp location result: RSRQ result out of range")
	ErrUERxTxTimeDiffOutOfRange          = errors.New("lpp location result: UE Rx-Tx time difference out of range")
	ErrSystemFrameNumberOutOfRange       = errors.New("lpp location result: system frame number out of range")
	ErrSystemFrameNumberInvalidBitLength = errors.New("lpp location result: system frame number must contain exactly 10 bits")
	ErrRSRPResultEncode                  = errors.New("lpp location result: RSRP result encode failed")
	ErrRSRPResultDecode                  = errors.New("lpp location result: RSRP result decode failed")
	ErrRSRQResultEncode                  = errors.New("lpp location result: RSRQ result encode failed")
	ErrRSRQResultDecode                  = errors.New("lpp location result: RSRQ result decode failed")
	ErrUERxTxTimeDiffEncode              = errors.New("lpp location result: UE Rx-Tx time difference encode failed")
	ErrUERxTxTimeDiffDecode              = errors.New("lpp location result: UE Rx-Tx time difference decode failed")
	ErrSystemFrameNumberEncode           = errors.New("lpp location result: system frame number encode failed")
	ErrSystemFrameNumberDecode           = errors.New("lpp location result: system frame number decode failed")
)
