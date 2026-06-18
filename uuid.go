package ulid

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidUUIDLen  = errors.New("uuid must contain exactly 32 hex characters")
	ErrInvalidUUIDChar = errors.New("invalid hexadecimal character in uuid")
)

// Returns the underlying 16 bytes of a ULID as a 32 hexadecimal string formatted with
// hyphones as a native, lowercase UUID string.
func (u ULID) UUID() string {
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", u[0:4], u[4:6], u[6:8], u[8:10], u[10:16])
}

// UUID enables UUID string compatibility with ULIDs. Both UUIDs and ULIDs are 128-bit
// identifiers, so it is possible to simply use the same 16 bytes as either a UUID or a
// ULID. However, the string representations are different. This method can parse a
// 32 character UUID string with mixed case with or without hyphens to return a valid
// ULID data structure with the same bytes as the original UUID.
func UUID(s string) (ULID, error) {
	var dst [16]byte
	hc := 0

	// Parse exactly 32 hex characters, skipping over hyphens
	for i := 0; i < len(s); i++ {
		// Grab the next character (note it can be multi-byte)
		ch := s[i]
		if ch == '-' {
			continue
		}

		// Do not overflow the 32-hex character buffer
		if hc >= 32 {
			return Zero, ErrInvalidUUIDLen
		}

		// Map character to its 4-bit nibble value (handling mixed case)
		var nibble byte
		switch {
		case ch >= '0' && ch <= '9':
			nibble = ch - '0'
		case ch >= 'a' && ch <= 'f':
			nibble = ch - 'a' + 10
		case ch >= 'A' && ch <= 'F':
			nibble = ch - 'A' + 10
		default:
			return Zero, ErrInvalidUUIDChar
		}

		// Shift the existing nibble left by 4 bits and add the new nibble
		if hc%2 == 0 {
			dst[hc/2] = nibble << 4
		} else {
			dst[hc/2] |= nibble
		}
		hc++
	}

	if hc != 32 {
		return Zero, ErrInvalidUUIDLen
	}
	return ULID(dst), nil
}
