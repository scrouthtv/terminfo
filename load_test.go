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
		werr := filepath.Walk(dir, func(file string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if fi.IsDir() || !fileRE.MatchString(file[len(dir)+1:]) || fi.Mode()&os.ModeSymlink != 0 {
				return nil
			}
			terms[filepath.Base(file)] = file
			return nil
		})
		if werr != nil {
			t.Fatalf("could not walk directory, got: %v", werr)
		}
	}

	for term, file := range terms {
		if term == "xterm-old" {
			continue
		}

		err := os.Setenv("TERM", term)
		if err != nil {
			t.Fatalf("could not set TERM environment variable, got: %v", err)
		}

		// open
		ti, err := LoadFromEnv()
		if err != nil {
			t.Fatalf("term %s expected no error, got: %v", term, err)
		}

		// check the name was saved correctly
		if ti.File != file {
			t.Errorf("term %s should have file %s, got: %s", term, file, ti.File)
		}

		// check we have at least one name
		if len(ti.Names) < 1 {
			t.Errorf("term %s expected names to have at least one value", term)
		}
	}
}
