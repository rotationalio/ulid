package main

import (
	cryptorand "crypto/rand"
	"flag"
	"fmt"
	mathrand "math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.rtnl.ai/ulid"
)

const usageText = `Rotational ULID debugging utility
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
`

const (
	defaultms = "Mon Jan 02 15:04:05.999 MST 2006"
	rfc3339ms = "2006-01-02T15:04:05.000Z07:00"
)

var (
	num    int
	quick  bool
	mono   bool
	zero   bool
	format string
	local  bool
	path   bool
	help   bool
)

func main() {
	// Set command line flags
	// Generate Options
	flag.IntVar(&num, "num", 1, "")
	flag.IntVar(&num, "n", 1, "")
	flag.BoolVar(&quick, "quick", false, "")
	flag.BoolVar(&quick, "q", false, "")
	flag.BoolVar(&mono, "mono", false, "")
	flag.BoolVar(&mono, "m", false, "")
	flag.BoolVar(&zero, "zero", false, "")
	flag.BoolVar(&zero, "z", false, "")

	// Inspect Options
	flag.StringVar(&format, "format", "default", "")
	flag.StringVar(&format, "f", "default", "")
	flag.BoolVar(&local, "local", false, "")
	flag.BoolVar(&local, "l", false, "")
	flag.BoolVar(&path, "path", false, "")
	flag.BoolVar(&path, "p", false, "")

	// General Options
	flag.BoolVar(&help, "help", false, "")
	flag.BoolVar(&help, "h", false, "")

	// Parse command line flags
	flag.Parse()
	if help {
		usage()
		os.Exit(0)
	}

	switch flag.NArg() {
	case 0:
		generate()
	default:
		parse()
	}
}

func usage() {
	fmt.Fprint(os.Stderr, usageText)
}

func generate() {
	if num < 1 {
		fmt.Fprintf(os.Stderr, "invalid --num %d\n", num)
		os.Exit(1)
	}

	// Create entropy from options
	entropy := cryptorand.Reader
	if quick {
		seed := time.Now().UnixNano()
		source := mathrand.NewSource(seed)
		entropy = mathrand.New(source)
	}
	if zero {
		entropy = zeroReader{}
	}
	if mono {
		entropy = ulid.Monotonic(entropy, 0)
	}

	// Generate ULIDs
	for i := 0; i < num; i++ {
		id, err := ulid.New(ulid.Timestamp(time.Now()), entropy)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		fmt.Fprintf(os.Stdout, "%s\n", id)
	}
}

func parse() {
	var formatFunc func(time.Time) string
	switch strings.ToLower(format) {
	case "default":
		formatFunc = func(t time.Time) string { return t.Format(defaultms) }
	case "rfc3339":
		formatFunc = func(t time.Time) string { return t.Format(rfc3339ms) }
	case "unix":
		formatFunc = func(t time.Time) string { return fmt.Sprint(t.Unix()) }
	case "ms":
		formatFunc = func(t time.Time) string { return fmt.Sprint(t.UnixNano() / 1e6) }
	default:
		fmt.Fprintf(os.Stderr, "invalid --format %s\n", format)
		os.Exit(1)
	}

	for _, s := range flag.Args() {
		if path {
			s = filepath.Base(s)
			s = strings.TrimSuffix(s, filepath.Ext(s))
		}

		id, err := ulid.Parse(s)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(1)
		}

		t := ulid.Time(id.Time())
		if !local {
			t = t.UTC()
		}
		fmt.Fprintf(os.Stderr, "%s\n", formatFunc(t))
	}
}

type zeroReader struct{}

func (zeroReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 0
	}
	return len(p), nil
}
