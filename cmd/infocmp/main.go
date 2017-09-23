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
	flagTerm  = flag.String("term", os.Getenv("TERM"), "term name")
	flagDebug = flag.Bool("debug", false, "debug")
)

func main() {
	flag.Parse()

	mask := "%s"
	if *flagDebug {
		mask = "%q"
	}

	var err error

	ti, err := terminfo.Load(*flagTerm)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("#\tReconstructed via %s from file: %s\n", strings.TrimPrefix(os.Args[0], "./"), ti.File)
	fmt.Printf("%s,\n", strings.TrimSpace(strings.Join(ti.Names, "|")))

	// print bool caps
	var boolCaps []string
	for i, b := range ti.Bools {
		if b {
			n := terminfo.BoolCapName(i)
			if ti.BoolsM[i] {
				n += "@"
			}
			boolCaps = append(boolCaps, n)
		}
	}
	sort.Strings(boolCaps)
	for _, c := range boolCaps {
		fmt.Printf("\t%s,\n", c)
	}

	// print num caps
	var numCaps []string
	for i, n := range ti.Nums {
		if n >= 0 {
			numCaps = append(numCaps, fmt.Sprintf("%s#%d", terminfo.NumCapName(i), n))
		} else if n == -2 {
			numCaps = append(numCaps, fmt.Sprintf("%s@", terminfo.NumCapName(i)))
		}
	}
	sort.Strings(numCaps)
	for _, n := range numCaps {
		fmt.Printf("\t%s,\n", n)
	}

	// print string caps
	stringCaps := make(map[string]string)
	var stringCapNames []string
	for i, n := range ti.Strings {
		z := terminfo.StringCapName(i)
		stringCapNames = append(stringCapNames, z)
		stringCaps[z] = n
	}
	for i := range ti.StringsM {
		z := terminfo.StringCapName(i)
		stringCapNames = append(stringCapNames, z)
		stringCaps[z] = "MISSING"
	}
	sort.Strings(stringCapNames)
	for _, n := range stringCapNames {
		v := stringCaps[n]
		if v == "MISSING" {
			fmt.Printf("\t%s@,\n", n)
		} else {
			fmt.Printf("\t%s="+mask+",\n", n, terminfo.Escape(v))
		}
	}
}
