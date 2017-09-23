package terminfo

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func TestLoad(t *testing.T) {
	var fileRE = regexp.MustCompile("^([0-9]+|[a-zA-Z])/")

	terms := make(map[string]string)
	for _, dir := range []string{"/lib/terminfo", "/usr/share/terminfo"} {
		filepath.Walk(dir, func(file string, fi os.FileInfo, err error) error {
			if fi.IsDir() || !fileRE.MatchString(file[len(dir)+1:]) || fi.Mode()&os.ModeSymlink != 0 {
				return nil
			}
			terms[filepath.Base(file)] = file
			return nil
		})
	}

	for term, file := range terms {
		//t.Logf("opening %s (%s)", file, term)

		os.Setenv("TERM", term)

		// open
		ti, err := LoadFromEnv()
		if err != nil {
			t.Fatalf("term %s expected no error, got: %v", term, err)
		}

		if ti.File != file {
			t.Errorf("term %s should have file %s, got: %s", term, file, ti.File)
		}

		// check we have at least one name
		if len(ti.Names) < 1 {
			t.Errorf("term %s expected names to have at least one value", term)
		}
		//t.Logf("term %s -- '%s'", term, ti.Names[len(ti.Names)-1])
	}
}
