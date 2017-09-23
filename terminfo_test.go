package terminfo

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func TestOpen(t *testing.T) {
	var fileRE = regexp.MustCompile("^([0-9]+|[a-zA-Z])/")

	for _, dir := range []string{"/lib/terminfo", "/usr/share/terminfo"} {
		t.Run(dir[1:], func(dir string) func(*testing.T) {
			return func(t *testing.T) {
				t.Parallel()
				//t.Logf("processing dir %s", dir)
				filepath.Walk(dir, func(file string, fi os.FileInfo, err error) error {
					if fi.IsDir() || !fileRE.MatchString(file[len(dir)+1:]) {
						//t.Logf("skipping: %s", file)
						return nil
					}

					term := filepath.Base(file)

					//t.Logf("opening %s (%s)", file, term)

					// open
					ti, err := Open(dir, term)
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

					return nil
				})
			}
		}(dir))
	}
}
