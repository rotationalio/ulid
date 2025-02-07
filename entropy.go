package ulid

import (
	"bufio"
	crand "crypto/rand"
	"encoding/binary"
	"io"
	"math"
	"math/bits"
	"math/rand"
	"sync"
	"time"
)

//===========================================================================
// Default Entropy
//===========================================================================

var defaultEntropy = func() io.Reader {
	return Pool(func() io.Reader {
		return Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)
	})
}()

// DefaultEntropy returns a thread-safe per process monotonically increasing
// entropy source. It uses a sync.Pool rather than a sync.Mutex to provide
// minimal contention for concurrent access.
func DefaultEntropy() io.Reader {
	return defaultEntropy
}

//===========================================================================
// Secure Entropy
//===========================================================================

var secureEntropy = func() io.Reader {
	return Pool(func() io.Reader {
		return Monotonic(crand.Reader, 0)
	})
}()

// SecureEntropy returns a thread-safe per process monotonically increasing
// entropy source that uses cryptographically random generation and a sync.Pool
func SecureEntropy() io.Reader {
	return secureEntropy
}

//===========================================================================
// Pool Entropy
//===========================================================================

// Provides a thread-safe source of entropy to assist with fast, concurrent access
// to random data generation. Specify the type of entropy to use

type PoolEntropy struct {
	sync.Pool
}

type MakeEntropy func() io.Reader

var _ io.Reader = &PoolEntropy{}

func Pool(entropy MakeEntropy) *PoolEntropy {
	return &PoolEntropy{
		Pool: sync.Pool{
			New: func() any { return entropy() },
		},
	}
}

func (e *PoolEntropy) Read(p []byte) (n int, err error) {
	r := e.Pool.Get().(io.Reader)
	n, err = r.Read(p)
	e.Pool.Put(r)
	return n, err
}

func (e *PoolEntropy) Get() io.Reader {
	return e.Pool.Get().(io.Reader)
}

func (e *PoolEntropy) Put(r io.Reader) {
	e.Pool.Put(r)
}

//===========================================================================
// Monotonic Readers
//===========================================================================

// MonotonicReader is an interface that should yield monotonically increasing
// entropy into the provided slice for all calls with the same ms parameter. If
// a MonotonicReader is provided to the New constructor, its MonotonicRead
// method will be used instead of Read.
type MonotonicReader interface {
	io.Reader
	MonotonicRead(ms uint64, p []byte) error
}

// Monotonic returns a source of entropy that yields strictly increasing entropy
// bytes, to a limit governeed by the `inc` parameter.
//
// Specifically, calls to MonotonicRead within the same ULID timestamp return
// entropy incremented by a random number between 1 and `inc` inclusive. If an
// increment results in entropy that would overflow available space,
// MonotonicRead returns ErrMonotonicOverflow.
//
// Passing `inc == 0` results in the reasonable default `math.MaxUint32`. Lower
// values of `inc` provide more monotonic entropy in a single millisecond, at
// the cost of easier "guessability" of generated ULIDs. If your code depends on
// ULIDs having secure entropy bytes, then it's recommended to use the secure
// default value of `inc == 0`, unless you know what you're doing.
//
// The provided entropy source must actually yield random bytes. Otherwise,
// monotonic reads are not guaranteed to terminate, since there isn't enough
// randomness to compute an increment number.
//
// The returned type isn't safe for concurrent use.
func Monotonic(entropy io.Reader, inc uint64) *MonotonicEntropy {
	m := MonotonicEntropy{
		Reader: bufio.NewReader(entropy),
		inc:    inc,
	}

	if m.inc == 0 {
		m.inc = math.MaxUint32
	}

	if rng, ok := entropy.(rng); ok {
		m.rng = rng
	}

	return &m
}

type rng interface{ Int63n(n int64) int64 }

// LockedMonotonicReader wraps a MonotonicReader with a sync.Mutex for safe
// concurrent use.
type LockedMonotonicReader struct {
	mu sync.Mutex
	MonotonicReader
}

// MonotonicRead synchronizes calls to the wrapped MonotonicReader.
func (r *LockedMonotonicReader) MonotonicRead(ms uint64, p []byte) (err error) {
	r.mu.Lock()
	err = r.MonotonicReader.MonotonicRead(ms, p)
	r.mu.Unlock()
	return err
}

// MonotonicEntropy is an opaque type that provides monotonic entropy.
type MonotonicEntropy struct {
	io.Reader
	ms      uint64
	inc     uint64
	entropy uint80
	rand    [8]byte
	rng     rng
}

// MonotonicRead implements the MonotonicReader interface.
func (m *MonotonicEntropy) MonotonicRead(ms uint64, entropy []byte) (err error) {
	if !m.entropy.IsZero() && m.ms == ms {
		err = m.increment()
		m.entropy.AppendTo(entropy)
	} else if _, err = io.ReadFull(m.Reader, entropy); err == nil {
		m.ms = ms
		m.entropy.SetBytes(entropy)
	}
	return err
}

// increment the previous entropy number with a random number
// of up to m.inc (inclusive).
func (m *MonotonicEntropy) increment() error {
	if inc, err := m.random(); err != nil {
		return err
	} else if m.entropy.Add(inc) {
		return ErrMonotonicOverflow
	}
	return nil
}

// random returns a uniform random value in [1, m.inc), reading entropy
// from m.Reader. When m.inc == 0 || m.inc == 1, it returns 1.
// Adapted from: https://golang.org/pkg/crypto/rand/#Int
func (m *MonotonicEntropy) random() (inc uint64, err error) {
	if m.inc <= 1 {
		return 1, nil
	}

	// Fast path for using a underlying rand.Rand directly.
	if m.rng != nil {
		// Range: [1, m.inc)
		return 1 + uint64(m.rng.Int63n(int64(m.inc))), nil
	}

	// bitLen is the maximum bit length needed to encode a value < m.inc.
	bitLen := bits.Len64(m.inc)

	// byteLen is the maximum byte length needed to encode a value < m.inc.
	byteLen := uint(bitLen+7) / 8

	// msbitLen is the number of bits in the most significant byte of m.inc-1.
	msbitLen := uint(bitLen % 8)
	if msbitLen == 0 {
		msbitLen = 8
	}

	for inc == 0 || inc >= m.inc {
		if _, err = io.ReadFull(m.Reader, m.rand[:byteLen]); err != nil {
			return 0, err
		}

		// Clear bits in the first byte to increase the probability
		// that the candidate is < m.inc.
		m.rand[0] &= uint8(int(1<<msbitLen) - 1)

		// Convert the read bytes into an uint64 with byteLen
		// Optimized unrolled loop.
		switch byteLen {
		case 1:
			inc = uint64(m.rand[0])
		case 2:
			inc = uint64(binary.LittleEndian.Uint16(m.rand[:2]))
		case 3, 4:
			inc = uint64(binary.LittleEndian.Uint32(m.rand[:4]))
		case 5, 6, 7, 8:
			inc = uint64(binary.LittleEndian.Uint64(m.rand[:8]))
		}
	}

	// Range: [1, m.inc)
	return 1 + inc, nil
}
