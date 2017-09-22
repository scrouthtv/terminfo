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
