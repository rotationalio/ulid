package ulid_test

import (
	"bytes"
	"crypto/rand"
	"errors"
	"testing"

	"go.rtnl.ai/ulid"
)

// ULID should be compatible with UUID strings.
func TestULID_UUID(t *testing.T) {
	// Fixtures of random ULIDs (timestamps between 2002-04-13 and 2026-07-13)
	// paired with their expected canonical UUID string representation.
	fixtures := []struct {
		ulid string
		uuid string
	}{
		{"01EYWM8MF35WYYY1PSVCC5JNEA", "0177b944-51e3-2f3d-ef06-d9db185955ca"},
		{"01K0XZMVRQG5THJA4G5VQPQ14K", "01983bfa-6f17-8175-1928-902eef6b8493"},
		{"00XYKTS7YZRKXXDDYK75ZDZMJJ", "00efa7ac-9fdf-c4fb-d6b7-d3397edfd252"},
		{"019A1EXF8SVA4TRHWRS84F2HGX", "014a82ee-bd19-da89-ac47-98ca08f1461d"},
		{"01GCZMY1Z59KB00T24TSSMHC74", "01833f4f-07e5-4cd6-0068-44d67348b0e4"},
		{"00ZXSSC27AWB0WMPXK6EYGJWBA", "00ff7396-08ea-e2c1-ca5b-b333bd09716a"},
		{"012RG925BN9XFK17VJJ4X47NFE", "01162091-1575-4f5f-309f-72913a43d5ee"},
		{"014E3QK930P30K1HFHG5PHP2HY", "01238779-a460-b0c1-30c5-f1816d1b0a3e"},
		{"01FG92CX2WAH3JVB6QWZT9KAK4", "017c1226-745c-5447-2dac-d7e7f499aa64"},
		{"011VXVE5MYZ7N20VCVQAHMWM6S", "010efbb7-169e-f9ea-206d-9bbaa34e50d9"},
		{"012VPKMSMAXHNNGXEAVMT7JYTB", "0116ed3a-668a-ec6b-5875-cadd34797b4b"},
		{"018WW9GED62AWEART99PN3FHM9", "01473898-39a6-12b8-e563-494daa37c689"},
		{"00ZEY519BWPN24ZTM3R98KGXEA", "00fbbc50-a57c-b544-4fea-83c2513875ca"},
		{"00ZRPXBAR1G06S1GRCVSQBN4EA", "00fe2dd5-ab01-800d-90c3-0cde6eba91ca"},
		{"01JDSQYZXTY6V5B5SEXX5GE4J1", "0193737f-7fba-f1b6-5597-2eef4b071241"},
		{"01HQ98HFEZVETT0H23G7RZ90KA", "018dd288-bddf-dbb5-a044-4381f1f4826a"},
		{"012SG9Z5R09XYV7GD4B4HYVSDM", "0116609f-9700-4f7d-b3c1-a45923ede5b4"},
		{"01FNW6HGEQSXHN0MR67DYJGZWN", "017d7868-c1d7-cf63-5053-063b7d287f95"},
		{"01E86FQKQJRVBDTWVXSG82EBYF", "01720cfb-cef2-c6d6-dd73-7dcc10272fcf"},
		{"01FWJ6J8KJVQZHS3P99GJR8HEA", "017f2469-2272-ddff-1c8e-c94c258445ca"},
		{"012CJRD59CGH5PWHN903V9NP1W", "01132586-952c-844b-6e46-a900f69ad83c"},
		{"015Z978SPW6DVSQ3CSXTJZA9K0", "012fd274-66dc-3377-9b8d-99eea5f52660"},
		{"016EB7YJAAB0730KTRM2NBQ0G4", "0133967f-494a-580e-304f-58a0aabb8204"},
		{"012R6F4DP4W8KR4ZKFN771EZ0V", "01160cf2-36c4-e227-827e-6fa9ce177c1b"},
		{"01ADJ5SPWNZDPWQF9E7K26CXQF", "0153645c-db95-fb6d-cbbd-2e3cc46676ef"},
		{"01ECXEBGZ9NEP29E3XY99KV3WW", "01733ae5-c3e9-abac-24b8-7df2533d8f9c"},
		{"01KNS7STBHVMJ89B8AJVZNDAEQ", "019d727c-e971-dd24-84ad-0a96ff56a9d7"},
		{"018B13D0W9GKEGTAR40CCHXK85", "0142c236-8389-84dd-0d2b-0403191ecd05"},
		{"01DAHHHWE4ZBJSJP6EWB88X2DE", "016aa318-f1c4-fae5-9958-cee2d08e89ae"},
		{"01AJ9BAEG2P3S88P31A11RV874", "015492b5-3a02-b0f2-8458-6150438da0e4"},
		{"01A8YE7VHJ4THNXQB558XW1NS9", "01523ce3-ee32-26a3-5edd-652a3bc0d729"},
		{"010YJGXYXQ6PF7Q0E9T8Z8J8MY", "0107a50e-fbb7-359e-7b81-c9d23e89229e"},
	}

	for _, tc := range fixtures {
		id := ulid.MustParse(tc.ulid)

		// ULID -> UUID string should match the canonical representation.
		if got := id.UUID(); got != tc.uuid {
			t.Errorf("%s: UUID() = %q, want %q", tc.ulid, got, tc.uuid)
		}

		// UUID string -> ULID should yield the same underlying bytes.
		parsed, err := ulid.UUID(tc.uuid)
		if err != nil {
			t.Fatalf("%s: UUID(%q) returned error: %v", tc.ulid, tc.uuid, err)
		}
		if !bytes.Equal(parsed[:], id[:]) {
			t.Errorf("%s: UUID(%q) = %s, want %s", tc.ulid, tc.uuid, parsed.String(), id.String())
		}
	}
}

func TestUUID(t *testing.T) {
	t.Parallel()
	expected := ulid.MustParse("00ZXSSC27AWB0WMPXK6EYGJWBA")

	tf := func(s string) func(t *testing.T) {
		return func(t *testing.T) {
			t.Parallel()
			parsed, err := ulid.UUID(s)
			if err != nil {
				t.Fatalf("UUID(%q) returned error: %v", s, err)
			}
			if !bytes.Equal(parsed[:], expected[:]) {
				t.Errorf("UUID(%q) = got %s, want %s", s, parsed.String(), expected.String())
			}
		}
	}

	t.Run("Lowercase", func(t *testing.T) {
		t.Parallel()

		t.Run("Hyphens", tf("00ff7396-08ea-e2c1-ca5b-b333bd09716a"))
		t.Run("NoHyphens", tf("00ff739608eae2c1ca5bb333bd09716a"))
	})

	t.Run("MixedCase", func(t *testing.T) {
		t.Parallel()
		t.Run("Hyphens", tf("00Ff7396-08Ea-E2c1-cA5b-b333Bd09716a"))
		t.Run("NoHyphens", tf("00Ff739608EaE2c1cA5bb333Bd09716a"))
	})

	t.Run("Uppercase", func(t *testing.T) {
		t.Parallel()
		t.Run("Hyphens", tf("00FF7396-08EA-E2C1-CA5B-B333BD09716A"))
		t.Run("NoHyphens", tf("00FF739608EAE2C1CA5BB333BD09716A"))
	})

	t.Run("Invalid", func(t *testing.T) {
		t.Parallel()

		tf := func(s string, err error) func(t *testing.T) {
			return func(t *testing.T) {
				t.Parallel()
				_, perr := ulid.UUID(s)
				if !errors.Is(perr, err) {
					t.Errorf("UUID(%q) returned error: %v, want %v", s, perr, err)
				}
			}
		}

		t.Run("Empty", tf("", ulid.ErrInvalidUUIDLen))
		t.Run("TooShort", tf("00FF7396-08EA-E2C1-CA5B-B333BD09716", ulid.ErrInvalidUUIDLen))
		t.Run("TooLong", tf("00FF7396-08EA-E2C1-CA5B-CE23-B333BD09716A", ulid.ErrInvalidUUIDLen))
		t.Run("InvalidChar", tf("00FF7396-08EA-E2C1-CA5B-B333BD09716X", ulid.ErrInvalidUUIDChar))
	})
}

func TestUUID_Fuzz(t *testing.T) {
	n := 1024
	if testing.Short() {
		n = 32
	}

	for i := 0; i < n; i++ {
		data := make([]byte, 16)
		if _, err := rand.Read(data); err != nil {
			t.Fatalf("rand.Read(%d) returned error: %v", len(data), err)
		}

		uu, err := ulid.Parse(data)
		if err != nil {
			t.Fatalf("Parse(%v) returned error: %v", data, err)
		}

		us := uu.UUID()
		parsed, err := ulid.UUID(us)
		if err != nil {
			t.Fatalf("UUID(%q) returned error: %v", us, err)
		}
		if !bytes.Equal(parsed[:], uu[:]) {
			t.Errorf("UUID(%q) = got %s, want %s", us, parsed.String(), uu.String())
		}
	}
}
