# Universally Unique Lexicographically Sortable Identifier

[![Project status](https://img.shields.io/github/release/rotationalio/ulid.svg?style=flat-square)](https://github.com/rotationalio/ulid/releases/latest)
![Build Status](https://github.com/rotationalio/ulid/actions/workflows/test.yaml/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/go.rtnl.ai/ulid)](https://goreportcard.com/report/go.rtnl.ai/ulid)
[![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/go.rtnl.ai/ulid)
[![Apache 2 licensed](https://img.shields.io/badge/license-Apache2-blue.svg)](https://raw.githubusercontent.com/oklog/ulid/master/LICENSE)

This go module is a derivative work of [github.com/oklog/ulid](https://github.com/oklog/ulid), created under the terms of the [Apache 2](LICENSE) license. We created this port of the original package because we found ourselves wrapping the package with helper functionality that didn't seem useful to anyone but us. This package is intended for a Rotational audience; if you need ULIDs we do recommend that you use the original package.

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

The CLI tool helps debug and generate ULIDs for your development workflow. Install the CLI using `go` as follows:

```shell
go install go.rtnl.ai/ulid/cmd/ulid@latest
```

Usage:

```shell
Rotational ULID debugging utility
Usage: generate or inspect a ULID

Generate:

    ulid [options]

    -n INT, --num INT     number of ULIDs to generate
    -q, --quick           use quick entropy (not cryptographic)
    -m, --mono            use monotonic entropy (for more than one ULID)
    -z, --zero            use zero entropy

Inspect:

    ulid [options] ULID [ULID ...]

    -f, --format string   time format (default, rfc3339, unix, ms)
    -l, --local           use local time instead of UTC
    -p, --path            assumes argument is a path with a ULID filename (strips directory and extension)

Options:

    -h, --help            display this help and exit
```

Examples:

```
$ ulid
01JKEHMRSH3HXYCYYZ1HZR2JBS
```

```
$ ulid -n 3 -mono
01JKEHNQPA0END3NHMFKB2Y6SE
01JKEHNQPA0END3NHMFNPBB9WE
01JKEHNQPA0END3NHMFRMCX384
```

```
$ ulid 01JKEHNQPA0END3NHMFKB2Y6SE
Thu Feb 06 21:11:53.29 UTC 2025
```

```
$ ulid -f rfc3339 --local 01JKEHNQPA0END3NHMFKB2Y6SE 01JKEHNQPA0END3NHMFNPBB9WE
2025-02-06T15:11:53.290-06:00
2025-02-06T15:11:53.290-06:00
```

```
$ ulid --path path/to/01JKEHNQPA0END3NHMFKB2Y6SE.json
Thu Feb 06 21:11:53.29 UTC 2025
```

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

On an Apple M1 Max, MacOS 15.3 and Go 1.23.3

```
goos: darwin
goarch: arm64
pkg: go.rtnl.ai/ulid
cpu: Apple M1 Max
BenchmarkNew/WithCrypoEntropy-10         	 9962818	      109.0 ns/op	      16 B/op	       1 allocs/op
BenchmarkNew/WithEntropy-10              	39486076	       33.55 ns/op	      16 B/op	       1 allocs/op
BenchmarkNew/WithoutEntropy-10              72576985	       16.62 ns/op	      16 B/op	       1 allocs/op
BenchmarkMustNew/WithCrypoEntropy-10        11441258	      107.4 ns/op	      16 B/op	       1 allocs/op
BenchmarkMustNew/WithEntropy-10             37700085	       31.30 ns/op	      16 B/op	       1 allocs/op
BenchmarkMustNew/WithoutEntropy-10          70010307	       18.37 ns/op	      16 B/op	       1 allocs/op
```

```
goos: darwin
goarch: arm64
pkg: go.rtnl.ai/ulid
cpu: Apple M1 Max
BenchmarkParse-10                          100000000	       10.65 ns/op	      2441.46 MB/s	       0 B/op	       0 allocs/op
BenchmarkParseStrict-10                     73864335	       15.97 ns/op	      1627.67 MB/s	       0 B/op	       0 allocs/op
BenchmarkMustParse-10                       95626101	       12.61 ns/op	      2061.36 MB/s	       0 B/op	       0 allocs/op
BenchmarkString-10                          86481555	       13.67 ns/op	      1170.76 MB/s	       0 B/op	       0 allocs/op
BenchmarkMarshal/Text-10                    94831988	       12.63 ns/op	      1266.42 MB/s	       0 B/op	       0 allocs/op
BenchmarkMarshal/TextTo-10                 100000000	       10.98 ns/op	      1456.62 MB/s	       0 B/op	       0 allocs/op
BenchmarkMarshal/Binary-10                 455631534	        2.760 ns/op	      5797.31 MB/s	       0 B/op	       0 allocs/op
BenchmarkMarshal/BinaryTo-10              1000000000	        1.111 ns/op	      14402.78 MB/s	       0 B/op	       0 allocs/op
BenchmarkUnmarshal/Text-10                 100000000	       10.43 ns/op	      2492.96 MB/s	       0 B/op	       0 allocs/op
BenchmarkUnmarshal/Binary-10               569854686	        2.135 ns/op	      7493.13 MB/s	       0 B/op	       0 allocs/op
BenchmarkNow-10                             31315614	       38.77 ns/op	       206.35 MB/s	       0 B/op	       0 allocs/op
BenchmarkTimestamp-10                     1000000000	        0.7833 ns/op	 10212.69 MB/s	       0 B/op	       0 allocs/op
BenchmarkTime-10                          1000000000	        0.8018 ns/op      9977.95 MB/s	       0 B/op	       0 allocs/op
BenchmarkSetTime-10                        950735085	        1.262 ns/op	      6338.36 MB/s	       0 B/op	       0 allocs/op
BenchmarkEntropy-10                        574565655	        2.042 ns/op	      4896.79 MB/s	       0 B/op	       0 allocs/op
BenchmarkSetEntropy-10                    1000000000	        0.9536 ns/op	 10486.70 MB/s	       0 B/op	       0 allocs/op
BenchmarkCompare-10                        522322389	        2.254 ns/op	     14197.26 MB/s	       0 B/op	       0 allocs/op
```

```
goos: darwin
goarch: arm64
pkg: go.rtnl.ai/ulid
cpu: Apple M1 Max
BenchmarkNew/WithMonotonicEntropy_SameTimestamp_Inc0-10         	        39891241	        29.24 ns/op	      16 B/op	       1 allocs/op
BenchmarkNew/WithMonotonicEntropy_DifferentTimestamp_Inc0-10    	        32685003	        36.05 ns/op	      16 B/op	       1 allocs/op
BenchmarkNew/WithMonotonicEntropy_SameTimestamp_Inc1-10         	        45091450	        25.16 ns/op	      16 B/op	       1 allocs/op
BenchmarkNew/WithMonotonicEntropy_DifferentTimestamp_Inc1-10    	        34196192	        35.31 ns/op	      16 B/op	       1 allocs/op
BenchmarkNew/WithCryptoMonotonicEntropy_SameTimestamp_Inc1-10   	        47389621	        25.40 ns/op	      16 B/op	       1 allocs/op
BenchmarkNew/WithCryptoMonotonicEntropy_DifferentTimestamp_Inc1-10         	39461244	        30.50 ns/op	      16 B/op	       1 allocs/op
BenchmarkMustNew/WithMonotonicEntropy_SameTimestamp_Inc0-10                	41440399	        29.14 ns/op	      16 B/op	       1 allocs/op
BenchmarkMustNew/WithMonotonicEntropy_DifferentTimestamp_Inc0-10           	32740442	        36.39 ns/op	      16 B/op	       1 allocs/op
BenchmarkMustNew/WithMonotonicEntropy_SameTimestamp_Inc1-10                	46801796	        26.14 ns/op	      16 B/op	       1 allocs/op
BenchmarkMustNew/WithMonotonicEntropy_DifferentTimestamp_Inc1-10           	32244736	        37.13 ns/op	      16 B/op	       1 allocs/op
BenchmarkMustNew/WithCryptoMonotonicEntropy_SameTimestamp_Inc1-10          	45454687	        26.91 ns/op	      16 B/op	       1 allocs/op
BenchmarkMustNew/WithCryptoMonotonicEntropy_DifferentTimestamp_Inc1-10     	36388584	        33.81 ns/op	      16 B/op	       1 allocs/op
```

## References

- [github.com/oklog/ulid](https://github.com/oklog/ulid)
