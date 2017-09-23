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
func decodeInt16(buf []byte) int16 {
	return int16(buf[1])<<8 | int16(buf[0])
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

// makemap converts a string slice to a map.
func makemap(s []string) map[string]int {
	m := make(map[string]int, len(s))
	for k, v := range s {
		m[v] = k
	}
	return m
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

	return int(decodeInt16(buf)), nil
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
		v := int(decodeInt16(buf[i*2 : i*2+2]))
		//log.Printf(">> %d: %s: %d", i, numCapNames[2*i], v)
		nums[i] = v
		if v == -2 {
			numsM[v] = true
		}
	}

	return nums, numsM, nil
}

// readStrings reads the next n strings and processes the string data table
// having length len.
func (d *decoder) readStrings(n, len int) (map[int]string, map[int]bool, error) {
	buf, err := d.readBytes(n * 2)
	if err != nil {
		return nil, nil, err
	}

	// load string table
	data, err := d.readBytes(len)
	if err != nil {
		return nil, nil, err
	}

	d.align()

	// process string data table
	s, m := make(map[int]string), make(map[int]bool)
	for i := 0; i < n; i++ {
		start := int(decodeInt16(buf[i*2 : i*2+2]))
		if start == -2 {
			m[i] = true
		} else if start >= 0 {
			if end := findNull(data[start:]); end != -1 {
				s[i] = string(data[start : start+end])
			} else {
				return nil, nil, ErrInvalidStringTable
			}
		}
	}

	return s, m, nil
}

// esc is the map of escape strings.
//var esc = map[rune]string{
//	'\001': `^A`,
//	'\002': `^B`,
//	'\003': `^C`,
//	'\004': `^D`,
//	'\005': `^E`,
//	'\006': `^F`,
//	'\a':   `^G`,
//	'\b':   `^H`,
//	'\t':   `^I`,
//	'\n':   `^J`,
//	'\013': `^K`,
//	'\014': `^L`,
//	'\r':   `^M`,
//	'\016': `^N`,
//	'\017': `^O`,
//	'\022': `^R`,
//	'\024': `^T`,
//	'\027': `^W`,
//	'\030': `^X`,
//	'\031': `^Y`,
//	'\032': `^Z`,
//	'\036': `^^`,
//
//	// commas are special characters in the infocmp file format.
//	',': `\054`,
//}

// escCh is a map of special chars to escape after an escape code.
/*var escCh = map[rune]string{
	//':': true,
	//'!': true,
	'\r': `\r`,
	'\n': `\n`,
}*/

func peek(rs []rune, pos, len int) rune {
	if pos < len {
		return rs[pos]
	}
	return 0
}

// Escape escapes a string using infocmp style escape codes.
func Escape(s string) string {
	rs := []rune(s)
	l := len(rs)
	var z string
	var p rune
	var afterEsc bool
	for i := 0; i < len(rs); i++ {
		r, n := rs[i], peek(rs, i+1, l)
		switch {
		case r == 0 || r == '\ufffd':
			z += `\0`

		case r == '\033':
			afterEsc = true
			z += `\E`

		case r == '\r' && n == '\n' && l > 2:
			z += string(`\r\n`)
			i++

		case r == '\r' /*&& i == l-1*/ && l > 2 && i != 0:
			z += string(`\r`)

		/*case r == '\016' && l > 1 && i == l-1:
		z += `\` + fmt.Sprintf("%03o", int(r))*/

		case r < ' ' && (p == '\033' || !afterEsc) /*(l < 3 || i == 0 || i == l-1)*/ :
			z += "^" + string(r+'@')

		/*case (r == '\r' || r == '\017') && (l > 2 && (i == 0 || i == l-1)):
		z += `\r`*/

		case p == '%' && (r == ':' || r == '!'):
			z += string(r)

		case r == ',' || r == ':' || r == '!' || r == '^' || !unicode.IsPrint(r) || r >= 128:
			z += `\` + fmt.Sprintf("%03o", int(r))

		default:
			z += string(r)
		}
		p = r
	}
	return z

	/*var z string
	var p rune
	for _, r := range s {
		switch {
		case r == '\000' || r == '\ufffd':
			z += `\0`

		case r == '\033':
			z += `\E`

		case r == '\r':
			z += `\r`

		case r == '\n':
			z += `\n`

		case p == '%' && (r == ':' || r == '!'):
			z += string(r)

		case r == ':' || r == '!' || r == ',' || r == '^' || !unicode.IsPrint(r) || r >= 128:
			z += `\` + fmt.Sprintf("%03o", int(r))

		default:
			z += string(r)
		}
		p = r
	}

	return z*/
}
