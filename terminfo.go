// Package terminfo implements reading terminfo files in pure go.
package terminfo

import (
	"errors"
	"io/ioutil"
	"path"
	"strconv"
	"strings"
)

var (
	// ErrInvalidFileSize is the invalid file size error.
	ErrInvalidFileSize = errors.New("invalid file size")

	// ErrUnexpectedFileEnd is the unexpected file end error.
	ErrUnexpectedFileEnd = errors.New("unexpected file end")

	// ErrInvalidStringTable is the invalid string table error.
	ErrInvalidStringTable = errors.New("invalid string table")

	// ErrInvalidMagic is the invalid magic error.
	ErrInvalidMagic = errors.New("invalid magic")

	// ErrInvalidHeader is the invalid header error.
	ErrInvalidHeader = errors.New("invalid header")

	// ErrInvalidExtendedHeader is the invalid extended header error.
	ErrInvalidExtendedHeader = errors.New("invalid extended header")

	// ErrEmptyTermName is the empty term name error.
	ErrEmptyTermName = errors.New("empty term name")

	// ErrDatabaseDirectoryNotFound is the database directory not found error.
	ErrDatabaseDirectoryNotFound = errors.New("database directory not found")

	// ErrFileNotFound is the file not found error.
	ErrFileNotFound = errors.New("file not found")
)

// Terminfo describes a terminal's capabilities.
type Terminfo struct {
	// File is the original source file.
	File string

	// Names are the provided cap names.
	Names []string

	// Bools are the bool capabilities.
	Bools map[int]bool

	// BoolsM are the missing bool capabilities.
	BoolsM map[int]bool

	// Nums are the num capabilities.
	Nums map[int]int

	// NumsM are the missing num capabilities.
	NumsM map[int]bool

	// Strings are the string capabilities.
	Strings map[int]string

	// StringsM are the missing string capabilities.
	StringsM map[int]bool

	// ExtBools are the extended bool capabilities.
	ExtBools map[int]bool

	// ExtBoolsNames is the map of extended bool capabilities to their index.
	ExtBoolNames map[int]string

	// ExtNums are the extended num capabilities.
	ExtNums map[int]int

	// ExtNumsNames is the map of extended num capabilities to their index.
	ExtNumNames map[int]string

	// ExtStrings are the extended string capabilities.
	ExtStrings map[int]string

	// ExtStringsNames is the map of extended string capabilities to their index.
	ExtStringNames map[int]string
}

// Decode decodes the terminfo data contained in buf.
func Decode(buf []byte) (*Terminfo, error) {
	var err error

	if len(buf) >= 4096 {
		return nil, ErrInvalidFileSize
	}

	d := &decoder{
		buf: buf,
		len: len(buf),
	}

	// check magic
	m, err := d.readInt16()
	if err != nil {
		return nil, err
	}
	if m != magic {
		return nil, ErrInvalidMagic
	}

	// read header
	h, _, err := d.readNums(5)
	if err != nil {
		return nil, err
	}

	// check header
	if hasInvalidCaps(h) {
		return nil, ErrInvalidHeader
	}

	// check remaining length
	if d.len-d.pos < capLength(h) {
		return nil, ErrUnexpectedFileEnd
	}

	// read term names
	names, err := d.readBytes(h[fieldNameSize])
	if err != nil {
		return nil, err
	}

	// read bool capabilities
	bools, boolsM, err := d.readBools(h[fieldBoolCount])
	if err != nil {
		return nil, err
	}

	// read num capabilities
	nums, numsM, err := d.readNums(h[fieldNumCount])
	if err != nil {
		return nil, err
	}

	// read string capabilities
	strs, strsM, err := d.readStrings(h[fieldStringCount], h[fieldTableSize])
	if err != nil {
		return nil, err
	}

	ti := &Terminfo{
		Names:    strings.Split(strings.TrimRight(string(names), "\x00"), "|"),
		Bools:    bools,
		BoolsM:   boolsM,
		Nums:     nums,
		NumsM:    numsM,
		Strings:  strs,
		StringsM: strsM,
	}

	// at the end of file, so no extended capabilities
	if d.pos >= d.len {
		return ti, nil
	}

	// decode extended header
	eh, _, err := d.readNums(5)
	if err != nil {
		return nil, err
	}

	// check extended offset field
	if hasInvalidExtOffset(eh) {
		return nil, ErrInvalidExtendedHeader
	}

	// check extended lengths in extended header
	if d.len-d.pos != extCapLength(eh) {
		return nil, ErrInvalidExtendedHeader
	}

	// read extended bools
	ti.ExtBools, _, err = d.readBools(eh[fieldExtBoolCount])
	if err != nil {
		return nil, err
	}

	// read extended nums
	ti.ExtNums, _, err = d.readNums(eh[fieldExtNumCount])
	if err != nil {
		return nil, err
	}

	// read extended string table
	count := eh[fieldExtBoolCount] + eh[fieldExtNumCount] + 2*eh[fieldExtStringCount]
	s, _, err := d.readStrings(count, eh[fieldExtTableSize])
	if err != nil {
		return nil, err
	}
	s = s

	// grab extended string cap values
	/*ti.ExtStrings, s = s[:eh[fieldExtStringCount]], s[eh[fieldExtStringCount]:]

	// grab extended bool, num, string names
	ti.ExtBoolsNames, s = makemap(s[:eh[fieldExtBoolCount]]), s[eh[fieldExtBoolCount]:]
	ti.ExtNumsNames, s = makemap(s[:eh[fieldExtNumCount]]), s[eh[fieldExtNumCount]:]
	ti.ExtStringsNames = makemap(s[:eh[fieldExtStringCount]])*/

	return ti, nil
}

// Open reads the terminfo file name from the specified directory dir.
func Open(dir, name string) (*Terminfo, error) {
	var err error
	var buf []byte
	var filename string
	for _, f := range []string{
		path.Join(dir, name[0:1], name),
		path.Join(dir, strconv.FormatUint(uint64(name[0]), 16), name),
	} {
		buf, err = ioutil.ReadFile(f)
		if err == nil {
			filename = f
			break
		}
	}
	if buf == nil {
		return nil, ErrFileNotFound
	}

	// decode
	ti, err := Decode(buf)
	if err != nil {
		return nil, err
	}

	// save original file name
	ti.File = filename

	// add to cache
	termCache.Lock()
	for _, n := range ti.Names {
		termCache.db[n] = ti
	}
	termCache.Unlock()

	return ti, nil
}

// boolCaps returns all bool and extended capabilities using f to format the
// index key.
func (ti *Terminfo) boolCaps(f func(BoolCapType) string) map[string]bool {
	m := make(map[string]bool, len(ti.Bools)+len(ti.ExtBools))
	for k, v := range ti.Bools {
		m[f(BoolCapType(k))] = v
	}
	for k, v := range ti.ExtBools {
		m[ti.ExtBoolNames[k]] = v
	}
	return m
}

// BoolCaps returns all bool and extended capabilities.
func (ti *Terminfo) BoolCaps() map[string]bool {
	return ti.boolCaps(BoolCapName)
}

// BoolCaps returns all bool and extended capabilities, using the short name as
// the map index.
func (ti *Terminfo) BoolCapsShort() map[string]bool {
	return ti.boolCaps(BoolCapNameShort)
}

// numCaps returns all num and extended capabilities using f to format the
// index key.
func (ti *Terminfo) numCaps(f func(NumCapType) string) map[string]int {
	m := make(map[string]int, len(ti.Nums)+len(ti.ExtNums))
	for k, v := range ti.Nums {
		m[f(NumCapType(k))] = v
	}
	for k, v := range ti.ExtNums {
		m[ti.ExtNumNames[k]] = v
	}
	return m
}

// NumCaps returns all int and extended capabilities.
func (ti *Terminfo) NumCaps() map[string]int {
	return ti.numCaps(NumCapName)
}

// NumCaps returns all int and extended capabilities, using the short name as
// the map index.
func (ti *Terminfo) NumCapsShort() map[string]int {
	return ti.numCaps(NumCapNameShort)
}

// stringCaps returns all string and extended capabilities using f to format the
// index key.
func (ti *Terminfo) stringCaps(f func(StringCapType) string) map[string]string {
	m := make(map[string]string, len(ti.Strings)+len(ti.ExtStrings))
	for k, v := range ti.Strings {
		m[f(StringCapType(k))] = v
	}
	for k, v := range ti.ExtStrings {
		m[ti.ExtStringNames[k]] = v
	}
	return m
}

// StringCaps returns all string and extended capabilities.
func (ti *Terminfo) StringCaps() map[string]string {
	return ti.stringCaps(StringCapName)
}

// StringCaps returns all string and extended capabilities, using the short name as
// the map index.
func (ti *Terminfo) StringCapsShort() map[string]string {
	return ti.stringCaps(StringCapNameShort)
}

// Sprintf formats the string cap s, interpolating parameters p.
func (ti *Terminfo) Sprintf(s StringCapType, p ...interface{}) string {
	return Sprintf(ti.Strings[int(s)], p...)
}

// Goto returns a string suitable for addressing the cursor at the given
// row and column. The origin 0, 0 is in the upper left corner of the screen.
func (ti *Terminfo) Goto(row, col int) string {
	return ti.Sprintf(CursorAddress, row, col)
}

// Puts emits the string to the writer, but expands inline padding indications
// (of the form $<[delay]> where [delay] is msec) to a suitable number of
// padding characters (usually null bytes) based upon the supplied baud. At
// high baud rates, more padding characters will be inserted.
/*func (ti *Terminfo) Puts(w io.Writer, s string, lines, baud int) (int, error) {
	var err error
	for {
		start := strings.Index(s, "$<")
		if start == -1 {
			// most strings don't need padding, which is good news!
			return io.WriteString(w, s)
		}

		end := strings.Index(s, ">")
		if end == -1 {
			// unterminated... just emit bytes unadulterated.
			return io.WriteString(w, "$<"+s)
		}

		var c int
		c, err = io.WriteString(w, s[:start])
		if err != nil {
			return n + c, err
		}
		n += c

		s = s[start+2:]
		val := s[:end]
		s = s[end+1:]
		var ms int
		var dot, mandatory, asterisk bool
		unit := 1000
		for _, ch := range val {
			switch {
			case ch >= '0' && ch <= '9':
				ms = (ms * 10) + int(ch-'0')
				if dot {
					unit *= 10
				}
			case ch == '.' && !dot:
				dot = true
			case ch == '*' && !asterisk:
				ms *= lines
				asterisk = true
			case ch == '/':
				mandatory = true
			default:
				break
			}
		}

		z, pad := ((baud/8)/unit)*ms, ti.Strings[PadChar]
		b := make([]byte, len(pad)*z)
		for bp := copy(b, pad); bp < len(b); bp *= 2 {
			copy(b[bp:], b[:bp])
		}

		if (!ti.Bools[XonXoff] && baud > int(ti.Nums[PaddingBaudRate])) || mandatory {
			c, err = w.Write(b)
			if err != nil {
				return n + c, err
			}
			n += c
		}
	}

	return n, nil
}*/

// Color takes a foreground and background color and returns string that sets
// them for this terminal.
//
// TODO redo with styles integer
/*func (ti *Terminfo) Color(fg, bg int) (rv string) {
	maxColors := int(ti.Nums[MaxColors])

	// map bright colors to lower versions if the color table only holds 8.
	if maxColors == 8 {
		if fg > 7 && fg < 16 {
			fg -= 8
		}
		if bg > 7 && bg < 16 {
			bg -= 8
		}
	}

	if maxColors > fg && fg >= 0 {
		rv += ti.Parm(SetAForeground, fg)
	}

	if maxColors > bg && bg >= 0 {
		rv += ti.Parm(SetABackground, bg)
	}

	return
}*/
