# Universally Unique Lexicographically Sortable Identifier

This go module is a derivative work of [github.com/oklog/ulid](https://github.com/oklog/ulid), created under the terms of the [Apache 2](LICENSE). We created this port of the original package because we found ourselves wrapping the package with helper functionality that didn't seem useful to anyone but us. This package is intended for a Rotational audience; if you need ULIDs we do recommend that you use the original package.

## Install

This package requires Go modules.

```shell
go get go.rtnl.ai/ulid
```

## Usage

ULIDs are constructed from two things: a timestamp with millisecond precision,
and some random data.

Timestamps are modeled as uint64 values representing a Unix time in milliseconds.
They can be produced by passing a [time.Time](https://pkg.go.dev/time#Time) to
[ulid.Timestamp](https://pkg.go.dev/github.com/oklog/ulid/v2#Timestamp),
or by calling  [time.Time.UnixMilli](https://pkg.go.dev/time#Time.UnixMilli)
and converting the returned value to `uint64`.

Random data is taken from a provided [io.Reader](https://pkg.go.dev/io#Reader).
This design allows for greater flexibility when choosing trade-offs, but can be
a bit confusing to newcomers.

If you just want to generate a ULID and don't (yet) care about details like
performance, cryptographic security, etc., use the
[ulid.Make](https://pkg.go.dev/github.com/oklog/ulid/v2#Make) helper function.
This function calls [time.Now](https://pkg.go.dev/time#Now) to get a timestamp,
and uses a source of entropy which is process-global,
[pseudo-random](https://pkg.go.dev/math/rand), and
[monotonic](https://pkg.go.dev/github.com/oklog/ulid/v2#LockedMonotonicReader).

```go
fmt.Println(ulid.Make())
// 01G65Z755AFWAKHE12NY0CQ9FH
```

More advanced use cases should utilize
[ulid.New](https://pkg.go.dev/github.com/oklog/ulid/v2#New).

```go
entropy := rand.New(rand.NewSource(time.Now().UnixNano()))
ms := ulid.Timestamp(time.Now())
fmt.Println(ulid.New(ms, entropy))
// 01G65Z755AFWAKHE12NY0CQ9FH
```

Care should be taken when providing a source of entropy.

The above example utilizes [math/rand.Rand](https://pkg.go.dev/math/rand#Rand),
which is not safe for concurrent use by multiple goroutines. Consider
alternatives such as
[x/exp/rand](https://pkg.go.dev/golang.org/x/exp/rand#LockedSource).
Security-sensitive use cases should always use cryptographically secure entropy
provided by [crypto/rand](https://pkg.go.dev/crypto/rand).

Performance-sensitive use cases should avoid synchronization when generating
IDs. One option is to use a unique source of entropy for each concurrent
goroutine, which results in no lock contention, but cannot provide strong
guarantees about the random data, and does not provide monotonicity within a
given millisecond. One common performance optimization is to pool sources of
entropy using a [sync.Pool](https://pkg.go.dev/sync#Pool).

Monotonicity is a property that says each ULID is "bigger than" the previous
one. ULIDs are automatically monotonic, but only to millisecond precision. ULIDs
generated within the same millisecond are ordered by their random component,
which means they are by default un-ordered. You can use
[ulid.MonotonicEntropy](https://pkg.go.dev/github.com/oklog/ulid/v2#MonotonicEntropy) or
[ulid.LockedMonotonicEntropy](https://pkg.go.dev/github.com/oklog/ulid/v2#LockedMonotonicEntropy)
to create ULIDs that are monotonic within a given millisecond, with caveats. See
the documentation for details.

## CLI Tool

Coming Soon!

## Background

A GUID/UUID can be suboptimal for many use-cases because:

- It isn't the most character efficient way of encoding 128 bits
- UUID v1/v2 is impractical in many environments, as it requires access to a unique, stable MAC address
- UUID v3/v5 requires a unique seed and produces randomly distributed IDs, which can cause fragmentation in many data structures
- UUID v4 provides no other information than randomness which can cause fragmentation in many data structures

A ULID however:

- Is compatible with UUID/GUID's
- 1.21e+24 unique ULIDs per millisecond (1,208,925,819,614,629,174,706,176 to be exact)
- Lexicographically sortable
- Canonically encoded as a 26 character string, as opposed to the 36 character UUID
- Uses Crockford's base32 for better efficiency and readability (5 bits per character)
- Case insensitive
- No special characters (URL safe)
- Monotonic sort order (correctly detects and handles the same millisecond)

## Specification

Below is the current specification of ULID as implemented in this repository.

### Components

**Timestamp**
- 48 bits
- UNIX-time in milliseconds
- Won't run out of space till the year 10889 AD

**Entropy**
- 80 bits
- User defined entropy source.
- Monotonicity within the same millisecond with [`ulid.Monotonic`](https://godoc.org/github.com/oklog/ulid#Monotonic)

### Encoding

[Crockford's Base32](http://www.crockford.com/wrmg/base32.html) is used as shown.
This alphabet excludes the letters I, L, O, and U to avoid confusion and abuse.

```
0123456789ABCDEFGHJKMNPQRSTVWXYZ
```

### Binary Layout and Byte Order

The components are encoded as 16 octets. Each component is encoded with the Most Significant Byte first (network byte order).

```
0                   1                   2                   3
 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                      32_bit_uint_time_high                    |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|     16_bit_uint_time_low      |       16_bit_uint_random      |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                       32_bit_uint_random                      |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
|                       32_bit_uint_random                      |
+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
```

### String Representation

```
 01AN4Z07BY      79KA1307SR9X4MV3
|----------|    |----------------|
 Timestamp           Entropy
  10 chars           16 chars
   48bits             80bits
   base32             base32
```


## Test

```shell
go test ./...
```

## Benchmarks

## References

- [github.com/oklog/ulid](https://github.com/oklog/ulid)
