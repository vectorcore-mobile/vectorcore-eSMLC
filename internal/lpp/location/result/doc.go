// Package result provides the bounded root scalar codecs needed by the
// Release 9 ECID measured-result path.
//
// It supports PhysicalCellID (INTEGER (0..503)), root ARFCN-ValueEUTRA
// (INTEGER (0..65535)), and four standalone root optional scalars:
// RSRPResult (0..97), RSRQResult (0..34), UERxTxTimeDiff (0..4095), and the
// exact ten-bit SystemFrameNumber. Their UPER widths are 9, 16, 7, 6, 12, and
// 10 bits. Callers compose fields, so these helpers never add octet alignment.
// Root-only CellGlobalIdEUTRA-AndUTRA is supported with mandatory MCC, a
// two- or three-digit MNC, and exclusive fixed-width E-UTRA/UTRA identities.
// It uses LPP constrained digits, not TBCD, and rejects wrapper extensions.
// MeasuredResultsElement optional bitmaps, lists, extension additions, open
// types, and provide-side envelopes remain outside this package. The package
// also supports root-only MeasuredResultsElement: its five root optional fields
// preserve absence separately from valid zero values and extension-present=true
// is rejected. Result lists and higher-level provide-side wrappers are absent.
package result
