package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/zellyn/go6502/asm"
	"github.com/zellyn/go6502/asm/flavors"
	"github.com/zellyn/go6502/asm/flavors/merlin"
	"github.com/zellyn/go6502/asm/flavors/redbook"
	"github.com/zellyn/go6502/asm/flavors/scma"
	"github.com/zellyn/go6502/asm/ihex"
	"github.com/zellyn/go6502/asm/lines"
	"github.com/zellyn/go6502/asm/opcodes"
)

var flavorNames = []string{
	"merlin",
	"scma",
	"redbooka",
	"redbookb",
}

var infile = flag.String("in", "", "input file")
var outfile = flag.String("out", "", "output file")
var listfile = flag.String("listing", "", "listing file")
var format = flag.String("format", "binary", "output format: binary/ihex")
var fill = flag.Uint("fillbyte", 0x00, "byte value to use when filling gaps between assmebler output regions")
var prefix = flag.Int("prefix", -1, "length of prefix to skip past addresses and bytes, -1 to guess")
var sweet16 = flag.Bool("sw16", false, "assemble sweet16 opcodes")
var flavorName = flag.String("flavor", "", fmt.Sprintf("assemble flavor: %s", strings.Join(flavorNames, ",")))

func main() {
	flag.Parse()
	if *infile == "" {
		fmt.Fprintln(os.Stderr, "no input file specified")
		os.Exit(1)
	}
	if *outfile == "" {
		fmt.Fprintln(os.Stderr, "no output file specified")
		os.Exit(1)
	}
	if *format != "binary" && *format != "ihex" {
		fmt.Fprintf(os.Stderr, "format must be binary or ihex; got '%s'\n", *format)
		os.Exit(1)
	}
	if *fill > 0xff {
		fmt.Fprintf(os.Stderr, "fillbyte must be <= 255; got '%s'\n", *format)
		os.Exit(1)
	}

	if *flavorName == "" {
		fmt.Fprintln(os.Stderr, "no flavor specified")
		os.Exit(1)
	}

	var f flavors.F
	var set opcodes.Set
	if *sweet16 {
		set |= opcodes.SetSweet16
	}
	switch *flavorName {
	case "merlin":
		f = merlin.New(set)
	case "scma":
		f = scma.New(set)
	case "redbooka":
		f = redbook.NewRedbookA(set)
	case "redbookb":
		f = redbook.NewRedbookB(set)
	default:
		fmt.Fprintf(os.Stderr, "invalid flavor: %q\n", *flavorName)
		os.Exit(1)
	}

	var o lines.OsOpener
	a := asm.NewAssembler(f, o)

	p := *prefix
	if p < 0 {
		var err error
		p, err = lines.GuessFilePrefixSize(*infile, o)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error trying to determine prefix length for file '%s'", *infile, err)
			os.Exit(1)
		}
	}

	if err := a.AssembleWithPrefix(*infile, p); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	out, err := os.Create(*outfile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer func() {
		err := out.Close()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}()

	m, err := a.Membuf()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	switch *format {
	case "binary":
		p := m.Piece(byte(*fill))
		n, err := out.Write(p.Data)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		if n != len(p.Data) {
			fmt.Fprintf(os.Stderr, "Error writing to '%s': wrote %d of %d bytes", *outfile, n, len(p.Data))
			os.Exit(1)

		}
	case "ihex":
		w := ihex.NewWriter(out)
		for _, p := range m.Pieces() {
			if err := w.Write(p.Addr, p.Data); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}
		w.End()
	default:
		fmt.Fprintf(os.Stderr, "format must be binary or ihex; got '%s'\n", *format)
		os.Exit(1)
	}

	if *listfile != "" {
		list, err := os.Create(*listfile)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer func() {
			err := list.Close()
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		}()

		err = a.GenerateListing(list, 3)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error while generating %s: %s", *listfile, err)
			os.Exit(1)
		}
	}

}
