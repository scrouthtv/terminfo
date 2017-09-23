package terminfo

import (
	"testing"
)

func TestCapSizes(t *testing.T) {
	if capCountBool*2 != len(boolCapNames) {
		t.Fatalf("boolCapNames should have same length as twice boolCapCount")
	}
	if capCountNum*2 != len(numCapNames) {
		t.Fatalf("numCapNames should have same length as twice numCapCount")
	}
	if capCountString*2 != len(stringCapNames) {
		t.Fatalf("stringCapNames should have same length as twice stringCapCount")
	}
}

func TestCapNames(t *testing.T) {
	for i := BoolCapType(0); i < BoolCapType(capCountBool); i++ {
		n, s := BoolCapName(i), BoolCapNameShort(i)
		if n == "" {
			t.Errorf("Bool cap %d should have name", i)
		}
		if s == "" {
			t.Errorf("Bool cap %d should have short name", i)
		}
		if n == s {
			t.Errorf("Bool cap %d name and short name should not equal (%s==%s)", i, n, s)
		}
	}
	for i := NumCapType(0); i < NumCapType(capCountNum); i++ {
		n, s := NumCapName(i), NumCapNameShort(i)
		if n == "" {
			t.Errorf("Num cap %d should have name", i)
		}
		if s == "" {
			t.Errorf("Num cap %d should have short name", i)
		}
		if n == s && n != "lines" {
			t.Errorf("Num cap %d name and short name should not equal (%s==%s)", i, n, s)
		}
	}
	for i := StringCapType(0); i < StringCapType(capCountString); i++ {
		n, s := StringCapName(i), StringCapNameShort(i)
		if n == "" {
			t.Errorf("String cap %d should have name", i)
		}
		if s == "" {
			t.Errorf("String cap %d should have short name", i)
		}
		if n == s && n != "tone" && n != "pulse" {
			t.Errorf("String cap %d name and short name should not equal (%s==%s)", i, n, s)
		}
	}
}
