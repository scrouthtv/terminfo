// Application infocmp should have the same output as the standard Unix infocmp
// -1 -L output.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"unicode"

	"github.com/xo/terminfo"
)

var (
	flagTerm     = flag.String("term", os.Getenv("TERM"), "term name")
	flagExtended = flag.Bool("x", false, "extended options")
)

func main() {
	flag.Parse()

	ti, err := terminfo.Load(*flagTerm)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("#\tReconstructed via %s from file: %s\n", strings.TrimPrefix(os.Args[0], "./"), ti.File)
	fmt.Printf("%s,\n", strings.TrimSpace(strings.Join(ti.Names, "|")))

	process(ti.BoolCaps, ti.ExtBoolCaps, ti.BoolsM, terminfo.BoolCapName, nil)
	process(
		ti.NumCaps, ti.ExtNumCaps, ti.NumsM, terminfo.NumCapName,
		func(v interface{}) string { return fmt.Sprintf("#%d", v) },
	)
	process(
		ti.StringCaps, ti.ExtStringCaps, ti.StringsM, terminfo.StringCapName,
		func(v interface{}) string { return "=" + escape(v.(string)) },
	)
}

func process(x, y interface{}, m map[int]bool, name func(int) string, mask func(interface{}) string) {
	printIt(x, m, name, mask)
	if *flagExtended {
		printIt(y, nil, name, mask)
	}
}

// process walks the values in z, adding missing elements in m. a mask func can
// be provided to format the values in z.
func printIt(z interface{}, m map[int]bool, name func(int) string, mask func(interface{}) string) {
	var names []string
	x := make(map[string]string)
	switch v := z.(type) {
	case func() map[string]bool:
		for n, a := range v() {
			if !a {
				continue
			}
			var f string
			if mask != nil {
				f = mask(a)
			}
			x[n], names = f, append(names, n)
		}

	case func() map[string]int:
		for n, a := range v() {
			if a < 0 {
				continue
			}
			var f string
			if mask != nil {
				f = mask(a)
			}
			x[n], names = f, append(names, n)
		}

	case func() map[string][]byte:
		for n, a := range v() {
			if a == nil {
				continue
			}
			var f string
			if mask != nil {
				f = mask(string(a))
			}
			x[n], names = f, append(names, n)
		}
	}

	// add missing
	for i := range m {
		n := name(i)
		x[n], names = "@", append(names, n)
	}

	// sort and print
	sort.Strings(names)
	for _, n := range names {
		fmt.Printf("\t%s%s,\n", n, x[n])
	}
}

// peek peeks a byte.
func peek(bs []byte, pos, len int) byte {
	if pos < len {
		return bs[pos]
	}
	return 0
}

// escape escapes a string using infocmp style escape codes.
func escape(s string) string {
	bs := []byte(s)
	l := len(bs)

	var z string
	var p byte
	var afterEsc bool
	for i := 0; i < len(bs); i++ {
		b, n := bs[i], peek(bs, i+1, l)
		switch {
		case b == 0 || b == '\200':
			z += `\0`

		case b == '\033':
			afterEsc = true
			z += `\E`

		case b == '\r' && n == '\n' && l > 2:
			z += string(`\r\n`)
			i++

		case b == '\r' /*&& i == l-1*/ && l > 2 && i != 0:
			z += string(`\r`)

		/*case r == '\016' && l > 1 && i == l-1:
		z += `\` + fmt.Sprintf("%03o", int(r))*/

		case b < ' ' && (p == '\033' || !afterEsc) /*(l < 3 || i == 0 || i == l-1)*/ :
			z += "^" + string(b+'@')

		/*case (r == '\r' || r == '\017') && (l > 2 && (i == 0 || i == l-1)):
		z += `\r`*/

		case p == '%' && (b == ':' || b == '!'):
			z += string(b)

		case b == ',' || b == ':' || b == '!' || b == '^' || !unicode.IsPrint(rune(b)) || b >= 128:
			z += `\` + fmt.Sprintf("%03o", int(b))

		default:
			z += string(b)
		}
		p = b
	}

	return z
}
