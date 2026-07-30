package uper

import "fmt"

func constrainedWidth(lower, upper uint64) (int, error) {
	if lower > upper {
		return 0, ErrInvalidRange
	}
	rangeSize := upper - lower
	if rangeSize == ^uint64(0) {
		return 0, ErrIntegerWidth
	}
	cardinality := rangeSize + 1
	width := 0
	for cardinality > 1 {
		width++
		cardinality = (cardinality + 1) >> 1
	}
	return width, nil
}

func (w *Writer) WriteConstrainedWholeNumber(value, lower, upper uint64) error {
	width, err := constrainedWidth(lower, upper)
	if err != nil {
		return err
	}
	if value < lower {
		return fmt.Errorf("%w: %d < %d", ErrValueBelowLower, value, lower)
	}
	if value > upper {
		return fmt.Errorf("%w: %d > %d", ErrValueAboveUpper, value, upper)
	}
	return w.WriteBits(value-lower, width)
}

func (r *Reader) ReadConstrainedWholeNumber(lower, upper uint64) (uint64, error) {
	width, err := constrainedWidth(lower, upper)
	if err != nil {
		return 0, err
	}
	offset, err := r.ReadBits(width)
	if err != nil {
		return 0, err
	}
	if offset > upper-lower {
		return 0, fmt.Errorf("%w: offset %d", ErrDecodedOutOfRange, offset)
	}
	return lower + offset, nil
}

func (w *Writer) WriteBoolean(value bool) error { return w.WriteBit(value) }
func (r *Reader) ReadBoolean() (bool, error)    { return r.ReadBit() }

func (w *Writer) WriteRootEnumerated(index, optionCount uint64) error {
	if optionCount == 0 {
		return ErrInvalidEnumerated
	}
	if index >= optionCount {
		return fmt.Errorf("%w: %d of %d", ErrInvalidEnumerated, index, optionCount)
	}
	return w.WriteConstrainedWholeNumber(index, 0, optionCount-1)
}
func (r *Reader) ReadRootEnumerated(optionCount uint64) (uint64, error) {
	if optionCount == 0 {
		return 0, ErrInvalidEnumerated
	}
	value, err := r.ReadConstrainedWholeNumber(0, optionCount-1)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrInvalidEnumerated, err)
	}
	return value, nil
}

func (w *Writer) WriteRootChoiceIndex(index, alternativeCount uint64) error {
	if alternativeCount == 0 {
		return ErrInvalidChoice
	}
	if index >= alternativeCount {
		return fmt.Errorf("%w: %d of %d", ErrInvalidChoice, index, alternativeCount)
	}
	return w.WriteConstrainedWholeNumber(index, 0, alternativeCount-1)
}
func (r *Reader) ReadRootChoiceIndex(alternativeCount uint64) (uint64, error) {
	if alternativeCount == 0 {
		return 0, ErrInvalidChoice
	}
	value, err := r.ReadConstrainedWholeNumber(0, alternativeCount-1)
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrInvalidChoice, err)
	}
	return value, nil
}

func (w *Writer) WriteOptionalBitmap(bits []bool) error {
	for _, present := range bits {
		if err := w.WriteBit(present); err != nil {
			return err
		}
	}
	return nil
}
func (r *Reader) ReadOptionalBitmap(count int) ([]bool, error) {
	if count < 0 || count > r.Remaining() {
		return nil, fmt.Errorf("%w: %d", ErrInvalidBitmap, count)
	}
	bits := make([]bool, count)
	for i := range bits {
		value, err := r.ReadBit()
		if err != nil {
			return nil, err
		}
		bits[i] = value
	}
	return bits, nil
}

func (w *Writer) WriteExtensionPresent(present bool) error { return w.WriteBit(present) }
func (r *Reader) ReadExtensionPresent() (bool, error)      { return r.ReadBit() }
func RequireNoExtension(present bool) error {
	if present {
		return ErrExtensionUnsupported
	}
	return nil
}
