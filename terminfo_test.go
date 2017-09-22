package terminfo

import (
	"testing"
)

// TODO look at unibillium tests
func TestOpen(t *testing.T) {
	ti, err := LoadFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	ti = ti
	//t.Logf("%q", ti.ExtStrings["kUP7"])
	//t.Logf("%q", ti.Strings[FlashScreen])
	//	b := bytes.NewBuffer(nil)
	//	ti.Strings[PadChar] = "*"
	//ti.Puts(b, ti.Strings[FlashScreen], 1, 9600)
	//	t.Logf("%q", b.Bytes())
	//t.Logf("%q", ti.Color(1, 1))
}

/*func BenchmarkParm(b *testing.B) {
	ti, err := LoadFromEnv()
	if err != nil {
		b.Fatal(err)
	}
	var r string
	for i := 0; i < b.N; i++ {
		r = ti.Color(7, 5)
	}
	result = r
}*/
