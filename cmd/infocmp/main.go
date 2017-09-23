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

	"github.com/xo/terminfo"
)

var (
	flagTerm = flag.String("term", os.Getenv("TERM"), "term name")
)

func main() {
	flag.Parse()

	ti, err := terminfo.Load(*flagTerm)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("#\tReconstructed via %s from file: %s\n", strings.TrimPrefix(os.Args[0], "./"), ti.File)
	fmt.Printf("%s,\n", strings.TrimSpace(strings.Join(ti.Names, "|")))

	process(ti.Bools, ti.BoolsM, terminfo.BoolCapName, nil)
	process(ti.Nums, ti.NumsM, terminfo.NumCapName, func(v interface{}) string { return fmt.Sprintf("#%d", v) })
	process(ti.Strings, ti.StringsM, terminfo.StringCapName, func(v interface{}) string { return "=" + terminfo.Escape(v.(string)) })
}

// process walks the values in z, adding missing elements in m. name and mask
// funcs can be provided to retrieve the name and format the mapped values in z.
func process(z interface{}, m map[int]bool, name func(int) string, mask func(interface{}) string) {
	var names []string
	x := make(map[string]string)
	switch v := z.(type) {
	case map[int]bool:
		for i, a := range v {
			if !a {
				continue
			}
			n := name(i)
			var f string
			if mask != nil {
				f = mask(a)
			}
			x[n], names = f, append(names, n)
		}
	case map[int]int:
		for i, a := range v {
			if a < 0 {
				continue
			}
			n := name(i)
			var f string
			if mask != nil {
				f = mask(a)
			}
			x[n], names = f, append(names, n)
		}
	case map[int]string:
		for i, a := range v {
			n := name(i)
			var f string
			if mask != nil {
				f = mask(a)
			}
			x[n], names = f, append(names, n)
		}
	}
	for i, _ := range m {
		n := name(i)
		x[n], names = "@", append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		fmt.Printf("\t%s%s,\n", n, x[n])
	}
}
