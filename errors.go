package ulid

import "errors"

var (
	// Occurs when parsing or unmarshaling ULIDs with the wrong number of bytes.
	ErrDataSize = errors.New("ulid: bad data size when unmarshaling")

	// Occurs when parsing or unmarshaling ULIDs with invalid Base32 encodings.
	ErrInvalidCharacters = errors.New("ulid: bad data characters when unmarshaling")

	// Occurs when marshalling ULIDs to a buffer of insufficient size.
	ErrBufferSize = errors.New("ulid: bad buffer size when marshaling")

	// Occurs when constructing a ULID with a time that is larger than MaxTime.
	ErrBigTime = errors.New("ulid: time too big")

	// Occurs when unmarshaling a ULID whose first character is
	// larger than 7, thereby exceeding the valid bit depth of 128.
	ErrOverflow = errors.New("ulid: overflow when unmarshaling")

	// Returned by a Monotonic entropy source when incrementing the previous ULID's
	// entropy bytes would result in overflow.
	ErrMonotonicOverflow = errors.New("ulid: monotonic entropy overflow")

	// Occurs when the value passed to scan cannot be unmarshaled into the ULID.
	ErrScanValue = errors.New("ulid: source value must be a string or byte slice")
)
