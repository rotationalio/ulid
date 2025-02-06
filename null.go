package ulid

import (
	"bytes"
	"database/sql/driver"
	"encoding/json"
)

// NullULID can be used with database/sql to represent ULIDs that are nullable without
// having to perform a check if the ULID is zero (and to ensure that NULL is set).
type NullULID struct {
	Valid bool
	ULID  ULID
}

func (nu *NullULID) Scan(value interface{}) error {
	if value == nil {
		nu.ULID, nu.Valid = Null, false
		return nil
	}

	err := nu.ULID.Scan(value)
	if err != nil {
		nu.Valid = false
		return err
	}

	nu.Valid = true
	return nil
}

func (nu NullULID) Value() (driver.Value, error) {
	if !nu.Valid {
		return nil, nil
	}
	return nu.ULID.Value()
}

func (nu NullULID) MarshalBinary() ([]byte, error) {
	if nu.Valid {
		return nu.ULID[:], nil
	}
	return []byte(nil), nil
}

func (nu *NullULID) UnmarshalBinary(data []byte) error {
	if len(data) != 16 {
		return ErrDataSize
	}

	copy(nu.ULID[:], data)
	nu.Valid = true
	return nil
}

func (nu *NullULID) MarshalText() ([]byte, error) {
	if nu.Valid {
		return nu.ULID.MarshalText()
	}
	return jsonNull, nil
}

func (nu *NullULID) UnmarshalText(data []byte) error {
	err := nu.ULID.UnmarshalText(data)
	if err != nil {
		nu.Valid = false
		return err
	}

	nu.Valid = true
	return nil
}

var jsonNull = []byte("null")

func (nu *NullULID) MarshalJSON() ([]byte, error) {
	if nu.Valid {
		return json.Marshal(nu.ULID)
	}
	return jsonNull, nil
}

func (nu *NullULID) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, jsonNull) {
		// Valid null ULID
		*nu = NullULID{}
		return nil
	}

	err := json.Unmarshal(data, &nu.ULID)
	nu.Valid = err == nil
	return err
}
