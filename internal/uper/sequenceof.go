package uper

import (
	"errors"
	"fmt"
)

func sequenceOfBounds(minCount, maxCount int) error {
	if minCount < 0 || maxCount < minCount {
		return fmt.Errorf("%w: SIZE(%d..%d)", ErrInvalidSequenceOfBounds, minCount, maxCount)
	}
	if _, err := constrainedWidth(uint64(minCount), uint64(maxCount)); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidSequenceOfBounds, err)
	}
	return nil
}

// WriteSequenceOf writes a bounded non-extensible SEQUENCE OF. It writes the
// constrained count first, then calls encodeElement once for every element in
// ascending wire order. It does not buffer elements or add alignment. On
// callback failure, the count and any earlier elements remain written.
func (w *Writer) WriteSequenceOf(count, minCount, maxCount int, encodeElement func(index int, w *Writer) error) error {
	if err := sequenceOfBounds(minCount, maxCount); err != nil {
		return err
	}
	if count < minCount {
		return fmt.Errorf("%w: %d < %d", ErrSequenceOfCountBelow, count, minCount)
	}
	if count > maxCount {
		return fmt.Errorf("%w: %d > %d", ErrSequenceOfCountAbove, count, maxCount)
	}
	if count > 0 && encodeElement == nil {
		return ErrSequenceOfEncodeCallback
	}
	if err := w.WriteConstrainedWholeNumber(uint64(count), uint64(minCount), uint64(maxCount)); err != nil {
		return err
	}
	for i := 0; i < count; i++ {
		if err := encodeElement(i, w); err != nil {
			return fmt.Errorf("%w at index %d: %w", ErrSequenceOfElementEncode, i, err)
		}
	}
	return nil
}

// ReadSequenceOf reads a bounded non-extensible SEQUENCE OF and invokes
// decodeElement in ascending wire order. It creates no element slice. The
// reader remains at the position reached on callback failure; reads are not
// transactional and no callback is invoked after an error.
func (r *Reader) ReadSequenceOf(minCount, maxCount int, decodeElement func(index int, r *Reader) error) (int, error) {
	if err := sequenceOfBounds(minCount, maxCount); err != nil {
		return 0, err
	}
	count64, err := r.ReadConstrainedWholeNumber(uint64(minCount), uint64(maxCount))
	if err != nil {
		if errors.Is(err, ErrUnexpectedEOF) {
			return 0, fmt.Errorf("%w: %w", ErrSequenceOfTruncatedCount, err)
		}
		return 0, err
	}
	count := int(count64)
	if count < minCount {
		return 0, fmt.Errorf("%w: %d < %d", ErrSequenceOfCountBelow, count, minCount)
	}
	if count > maxCount {
		return 0, fmt.Errorf("%w: %d > %d", ErrSequenceOfCountAbove, count, maxCount)
	}
	if count > 0 && decodeElement == nil {
		return 0, ErrSequenceOfDecodeCallback
	}
	for i := 0; i < count; i++ {
		if err := decodeElement(i, r); err != nil {
			if errors.Is(err, ErrUnexpectedEOF) {
				return count, fmt.Errorf("%w at index %d: %w", ErrSequenceOfTruncatedElement, i, fmt.Errorf("%w: %w", ErrSequenceOfElementDecode, err))
			}
			return count, fmt.Errorf("%w at index %d: %w", ErrSequenceOfElementDecode, i, err)
		}
	}
	return count, nil
}
