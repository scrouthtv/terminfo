package terminfo

import (
	"fmt"
	"unicode"
)

// magic is the file magic for terminfo files.
const magic = 0432

// header fields.
const (
	fieldNameSize = iota
	fieldBoolCount
	fieldNumCount
	fieldStringCount
	fieldTableSize
)

// header extended fields.
const (
	fieldExtBoolCount = iota
	fieldExtNumCount
	fieldExtStringCount
	fieldExtOffsetCount
	fieldExtTableSize
)

// hasInvalidCaps returns determines if the capabilities in h are invalid.
func hasInvalidCaps(h map[int]int) bool {
	return h[fieldBoolCount] > CapCountBool ||
		h[fieldNumCount] > CapCountNum ||
		h[fieldStringCount] > CapCountString
}

// capLength returns the total length of the capabilities in bytes.
func capLength(h map[int]int) int {
	return h[fieldNameSize] +
		h[fieldBoolCount] +
		(h[fieldNameSize]+h[fieldBoolCount])%2 + // account for word align
		h[fieldNumCount]*2 +
		h[fieldStringCount]*2 +
		h[fieldTableSize]
}

// hasInvalidExtOffset determines if the extended offset field is valid.
func hasInvalidExtOffset(h map[int]int) bool {
	return h[fieldExtBoolCount]+
		h[fieldExtNumCount]+
		h[fieldExtStringCount]*2 != h[fieldExtOffsetCount]
}

// extCapLength returns the total length of extended capabilities in bytes.
func extCapLength(h map[int]int) int {
	return h[fieldExtBoolCount] +
		h[fieldExtBoolCount]%2 + // account for word align
		h[fieldExtNumCount]*2 +
		h[fieldExtOffsetCount]*2 +
		h[fieldExtTableSize]
}

// decodeInt16 decodes a 16-bit little endian integer in buf.
func decodeInt16(buf []byte) int {
	return int(int16(buf[1])<<8 | int16(buf[0]))
}

// findNull finds the position of null in buf.
func findNull(buf []byte) int {
	for i := 0; i < len(buf); i++ {
		if buf[i] == 0 {
			return i
		}
	}
	return -1
}

// decoder holds state info while decoding a terminfo file.
type decoder struct {
	buf []byte
	pos int
	len int
}

// align increments pos when at an uneven word boundary.
func (d *decoder) align() {
	if d.pos%2 == 1 {
		d.pos++
	}
}

// readBytes reads the next n bytes of buf, incrementing pos by n.
func (d *decoder) readBytes(n int) ([]byte, error) {
	if d.len < d.pos+n {
		return nil, ErrUnexpectedFileEnd
	}

	n, d.pos = d.pos, d.pos+n

	return d.buf[n:d.pos], nil
}

// readInt16 reads the next 16 bit integer.
func (d *decoder) readInt16() (int, error) {
	buf, err := d.readBytes(2)
	if err != nil {
		return 0, err
	}

	return decodeInt16(buf), nil
}

// readBools reads the next n bools.
func (d *decoder) readBools(n int) (map[int]bool, map[int]bool, error) {
	buf, err := d.readBytes(n)
	if err != nil {
		return nil, nil, err
	}

	d.align()

	// process
	bools, boolsM := make(map[int]bool), make(map[int]bool)
	for i, b := range buf {
		bools[i] = b == 1
		if int8(b) == -2 {
			boolsM[i] = true
		}
	}

	return bools, boolsM, nil
}

// readNums reads the next n nums.
func (d *decoder) readNums(n int) (map[int]int, map[int]bool, error) {
	buf, err := d.readBytes(n * 2)
	if err != nil {
		return nil, nil, err
	}

	// process
	nums, numsM := make(map[int]int), make(map[int]bool)
	for i := 0; i < n; i++ {
		v := decodeInt16(buf[i*2 : i*2+2])
		if v >= 0 {
			nums[i] = v
		} else if v == -2 {
			numsM[i] = true
		}
	}

	return nums, numsM, nil
}

// readStringTable reads the string data for n strings and the accompanying data
// table of length sz.
func (d *decoder) readStringTable(n, sz int) ([][]byte, []int, error) {
	buf, err := d.readBytes(n * 2)
	if err != nil {
		return nil, nil, err
	}

	// read string data table
	data, err := d.readBytes(sz)
	if err != nil {
		return nil, nil, err
	}

	d.align()

	// process
	s := make([][]byte, n)
	var m []int
	for i := 0; i < n; i++ {
		start := decodeInt16(buf[i*2 : i*2+2])
		if start == -2 {
			m = append(m, i)
		} else if start >= 0 {
			if end := findNull(data[start:]); end != -1 {
				s[i] = data[start : start+end]
			} else {
				return nil, nil, ErrInvalidStringTable
			}
		}
	}

	return s, m, nil
}

// readStrings reads the next n strings and processes the string data table of
// length sz.
func (d *decoder) readStrings(n, sz int) (map[int][]byte, map[int]bool, error) {
	s, m, err := d.readStringTable(n, sz)
	if err != nil {
		return nil, nil, err
	}

	strs := make(map[int][]byte)
	for k, v := range s {
		strs[k] = v
	}

	strsM := make(map[int]bool, len(m))
	for _, k := range m {
		strsM[k] = true
	}

	return strs, strsM, nil
}

// makemap converts a string slice to a map.
func makemap(s [][]byte) map[int]string {
	m := make(map[int]string, len(s))
	for k, v := range s {
		m[k] = string(v)
	}
	return m
}

// peek peeks a byte.
func peek(bs []byte, pos, len int) byte {
	if pos < len {
		return bs[pos]
	}
	return 0
}

// Escape escapes a string using infocmp style escape codes.
func Escape(s string) string {
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
