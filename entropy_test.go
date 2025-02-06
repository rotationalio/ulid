package ulid_test

import (
	"bytes"
	crand "crypto/rand"
	"fmt"
	"io"
	"math"
	"math/rand"
	"sync"
	"testing"
	"time"

	"go.rtnl.ai/ulid"
)

func TestPoolEntropy(t *testing.T) {
	wg := sync.WaitGroup{}
	entropy := ulid.Pool(func() io.Reader { return rand.New(rand.NewSource(time.Now().UnixNano())) })

	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 128; i++ {
				uu, err := ulid.New(ulid.Now(), entropy)
				if err != nil {
					t.Errorf("could not create ulid: %s", err)
				}

				if uu.IsZero() {
					t.Error("expected ulid to not be null")
				}
			}
		}()
	}

	wg.Wait()
}

func TestMonotonic(t *testing.T) {
	now := ulid.Now()
	for _, e := range []struct {
		name string
		mk   func() io.Reader
	}{
		{"cryptorand", func() io.Reader { return crand.Reader }},
		{"mathrand", func() io.Reader { return rand.New(rand.NewSource(int64(now))) }},
	} {
		for _, inc := range []uint64{
			0,
			1,
			2,
			math.MaxUint8 + 1,
			math.MaxUint16 + 1,
			math.MaxUint32 + 1,
		} {
			inc := inc
			entropy := ulid.Monotonic(e.mk(), uint64(inc))

			t.Run(fmt.Sprintf("entropy=%s/inc=%d", e.name, inc), func(t *testing.T) {
				t.Parallel()

				var prev ulid.ULID
				for i := 0; i < 10000; i++ {
					next, err := ulid.New(123, entropy)
					if err != nil {
						t.Fatal(err)
					}

					if prev.Compare(next) >= 0 {
						t.Fatalf("prev: %v %v > next: %v %v",
							prev.Time(), prev.Entropy(), next.Time(), next.Entropy())
					}

					prev = next
				}
			})
		}
	}
}

func TestMonotonicOverflow(t *testing.T) {
	t.Parallel()

	entropy := ulid.Monotonic(
		io.MultiReader(
			bytes.NewReader(bytes.Repeat([]byte{0xFF}, 10)), // Entropy for first ULID
			crand.Reader, // Following random entropy
		),
		0,
	)

	prev, err := ulid.New(0, entropy)
	if err != nil {
		t.Fatal(err)
	}

	next, err := ulid.New(prev.Time(), entropy)
	if have, want := err, ulid.ErrMonotonicOverflow; have != want {
		t.Errorf("have ulid: %v %v err: %v, want err: %v",
			next.Time(), next.Entropy(), have, want)
	}
}

func TestMonotonicSafe(t *testing.T) {
	t.Parallel()

	var (
		rng  = rand.New(rand.NewSource(time.Now().UnixNano()))
		safe = &ulid.LockedMonotonicReader{MonotonicReader: ulid.Monotonic(rng, 0)}
		t0   = ulid.Timestamp(time.Now())
	)

	errs := make(chan error, 100)
	for i := 0; i < cap(errs); i++ {
		go func() {
			u0 := ulid.MustNew(t0, safe)
			u1 := u0
			for j := 0; j < 1024; j++ {
				u0, u1 = u1, ulid.MustNew(t0, safe)
				if u0.String() >= u1.String() {
					errs <- fmt.Errorf(
						"%s (%d %x) >= %s (%d %x)",
						u0.String(), u0.Time(), u0.Entropy(),
						u1.String(), u1.Time(), u1.Entropy(),
					)
					return
				}
			}
			errs <- nil
		}()
	}

	for i := 0; i < cap(errs); i++ {
		if err := <-errs; err != nil {
			t.Fatal(err)
		}
	}
}
