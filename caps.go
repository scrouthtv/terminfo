package terminfo

//go:generate go run gen.go

// BoolCapType is the bool capabilities type.
type BoolCapType int

// NumCapType is the num capabilities type.
type NumCapType int

// StringCapType is the string capabilities type.
type StringCapType int

// BoolCapName returns the bool capability name.
func BoolCapName(i BoolCapType) string {
	return boolCapNames[2*int(i)]
}

// BoolCapNameShort returns the short bool capability name.
func BoolCapNameShort(i BoolCapType) string {
	return boolCapNames[2*int(i)+1]
}

// NumCapName returns the num capability name.
func NumCapName(i NumCapType) string {
	return numCapNames[2*int(i)]
}

// NumCapNameShort returns the short num capability name.
func NumCapNameShort(i NumCapType) string {
	return numCapNames[2*int(i)+1]
}

// StringCapName returns the string capability name.
func StringCapName(i StringCapType) string {
	return stringCapNames[2*int(i)]
}

// StringCapNameShort returns the short string capability name.
func StringCapNameShort(i StringCapType) string {
	return stringCapNames[2*int(i)+1]
}

// String satisfies Stringer interface.
func (i BoolCapType) String() string {
	if s := BoolCapNameShort(i); s != "" {
		return s
	}
	return BoolCapName(i)
}

// String satisfies Stringer interface.
func (i NumCapType) String() string {
	if s := NumCapNameShort(i); s != "" {
		return s
	}
	return NumCapName(i)
}

// String satisfies Stringer interface.
func (i StringCapType) String() string {
	if s := StringCapNameShort(i); s != "" {
		return s
	}
	return StringCapName(i)
}
