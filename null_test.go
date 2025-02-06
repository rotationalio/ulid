package ulid_test

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	. "go.rtnl.ai/ulid"
)

func TestNullULIDScan(t *testing.T) {
	var u ULID
	var nu NullULID

	uNilErr := u.Scan(nil)
	nuNilErr := nu.Scan(nil)
	if uNilErr != uNilErr {
		t.Fatalf("expected errors to be equal, got %s, %s", uNilErr, nuNilErr)
	}

	uInvalidStringErr := u.Scan("test")
	nuInvalidStringErr := nu.Scan("test")
	if uInvalidStringErr != nuInvalidStringErr {
		t.Fatalf("expected errors to be equal, got %s, %s", uInvalidStringErr, nuInvalidStringErr)
	}

	valid := "01HTNMW2JAW89YSBG7NFPHABA4"
	uValidErr := u.Scan(valid)
	nuValidErr := nu.Scan(valid)
	if uValidErr != nil || nuValidErr != nil {
		t.Fatalf("expected no errors, got %s, %s", uValidErr, nuValidErr)
	}
}

func TestNullULIDValue(t *testing.T) {
	var u ULID
	var nu NullULID

	nuValue, nuErr := nu.Value()
	if nuErr != nil {
		t.Fatal(nuErr)
	}
	if nuValue != nil {
		t.Fatalf("expected nil, got %s", nuValue)
	}

	u = MustParse("01HTNMW2JAW89YSBG7NFPHABA4")
	nu = NullULID{
		ULID:  MustParse("01HTNMW2JAW89YSBG7NFPHABA4"),
		Valid: true,
	}

	uValue, uErr := u.Value()
	if uErr != nil {
		t.Fatal(uErr)
	}

	nuValue, nuErr = nu.Value()
	if nuErr != nil {
		t.Fatal(nuErr)
	}

	if !reflect.DeepEqual(nuValue, uValue) {
		t.Fatalf("expected ulid %s and nullulid %s to be equal ", uValue, nuValue)
	}
}

func TestNullULIDMarshalText(t *testing.T) {
	tests := []struct {
		nullULID NullULID
	}{
		{
			nullULID: NullULID{},
		},
		{
			nullULID: NullULID{
				ULID:  MustParse("01HTNMW2JAW89YSBG7NFPHABA4"),
				Valid: true,
			},
		},
	}
	for _, test := range tests {
		var uText []byte
		var uErr error
		nuText, nuErr := test.nullULID.MarshalText()
		if test.nullULID.Valid {
			uText, uErr = test.nullULID.ULID.MarshalText()
		} else {
			uText = []byte("null")
		}

		if nuErr != uErr {
			t.Fatalf("expected error %e, got %e", nuErr, uErr)
		}

		if !bytes.Equal(nuText, uText) {
			t.Fatalf("expected text data %s, got %s", string(nuText), string(uText))
		}
	}
}

func TestNullULIDMarshalBinary(t *testing.T) {
	tests := []struct {
		nullULID NullULID
	}{
		{
			nullULID: NullULID{},
		},
		{
			nullULID: NullULID{
				ULID:  MustParse("01HTNMW2JAW89YSBG7NFPHABA4"),
				Valid: true,
			},
		},
	}
	for _, test := range tests {
		var uBinary []byte
		var uErr error
		nuBinary, nuErr := test.nullULID.MarshalBinary()
		if test.nullULID.Valid {
			uBinary, uErr = test.nullULID.ULID.MarshalBinary()
		} else {
			uBinary = []byte(nil)
		}

		if nuErr != uErr {
			t.Fatalf("expected error %e, got %e", nuErr, uErr)
		}

		if !bytes.Equal(nuBinary, uBinary) {
			t.Fatalf("expected binary data %s, got %s", string(nuBinary), string(uBinary))
		}
	}
}

func TestNullULIDMarshalJSON(t *testing.T) {
	jsonNull, _ := json.Marshal(nil)
	tests := []struct {
		nullULID    NullULID
		expected    []byte
		expectedErr error
	}{
		{
			nullULID:    NullULID{},
			expected:    jsonNull,
			expectedErr: nil,
		},
		{
			nullULID: NullULID{
				ULID:  MustParse("01HTNMW2JAW89YSBG7NFPHABA4"),
				Valid: true,
			},
			expected:    []byte(`"01HTNMW2JAW89YSBG7NFPHABA4"`),
			expectedErr: nil,
		},
	}
	for _, test := range tests {
		data, err := json.Marshal(&test.nullULID)

		if test.expectedErr != err {
			t.Fatalf("expected error %e, got %e", test.expectedErr, err)
		}

		if !bytes.Equal(data, test.expected) {
			t.Fatalf("expected binary data %s, got %s", string(test.expected), string(data))
		}
	}
}

func TestNullULIDUnmarshalJSON(t *testing.T) {
	jsonNull, _ := json.Marshal(nil)
	jsonULID, _ := json.Marshal(MustParse("01HTNMW2JAW89YSBG7NFPHABA4"))

	var nu NullULID
	if err := json.Unmarshal(jsonNull, &nu); err != nil {
		t.Fatal(err)
	}

	if nu.Valid {
		t.Fatal("expected invalid NullULID")
	}

	if err := json.Unmarshal(jsonULID, &nu); err != nil {
		t.Fatal(err)
	}

	if !nu.Valid {
		t.Fatal("expected valid NullULID")
	}
}
