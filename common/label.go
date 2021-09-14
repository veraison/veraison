package common

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// Label is attached to, and identifies a claim (or a set of claims). A label
// can be either an integer or a string.
type Label struct {
	intValue int
	strValue string
	isInt    bool
}

// Value returns the value of the label an interface{}. The underlying value
// will be either a string or an int.
func (l Label) Value() interface{} {
	if l.isInt {
		return l.intValue
	}

	return l.strValue
}

// IsInt returns true iff the label value is an integer.
func (l Label) IsInt() bool {
	return l.isInt
}

// AsInt returns the value of the label as an integer, along with an error. If
// the underlying label value is an integer, error will be nil. For string
// labels, the value of -1 is returned, and the error is set.
func (l Label) AsInt() (int, error) {
	if l.isInt {
		return l.intValue, nil
	}

	return -1, fmt.Errorf("label value not an integer")
}

// String returns the Label value as a string. In case of string values, the
// value itself is returned. For integer Labels, the integer is formatted as
// string (as with fmt).
func (l Label) String() string {
	if l.isInt {
		return fmt.Sprint(l.intValue)
	}

	return l.strValue
}

// FromString populates the label value from the specified string, preferring
// to interpret it as a string representation of an int. Note: this is
// different from StringLabel(), which always sets the string value, even if
// provided text is a representation of an int.
func (l *Label) FromString(text string) {
	intVal, err := strconv.Atoi(text)
	if err == nil {
		l.intValue = intVal
		l.isInt = true
	} else {
		l.strValue = text
		l.isInt = false
	}
}

func (l Label) MarshalJSON() ([]byte, error) {
	if l.isInt {
		return json.Marshal(fmt.Sprint(l.intValue))
	}

	return json.Marshal(l.strValue)
}

func (l *Label) UnmarshalJSON(data []byte) error {
	var strVal string
	if err := json.Unmarshal(data, &strVal); err != nil {
		return err
	}

	l.FromString(strVal)

	return nil
}

func (l Label) MarshalCBOR() ([]byte, error) {
	if l.isInt {
		return em.Marshal(l.intValue)
	}

	return em.Marshal(l.strValue)
}

func (l *Label) UnmarshalCBOR(data []byte) error {
	var intVal int
	err := dm.Unmarshal(data, &intVal)
	if err == nil {
		l.intValue = intVal
		l.isInt = true
		return nil
	}

	var strVal string
	err = dm.Unmarshal(data, &strVal)
	if err == nil {
		l.strValue = strVal
		l.isInt = false
	}

	return fmt.Errorf("could not extract Label from CBOR: %v", data)
}

// NewIntLabel creates a new Label from the specified int value.
func NewIntLabel(value int) Label {
	return Label{intValue: value, isInt: true}
}

// NewStringLabel creates a new Label from the specified string value.
func NewStringLabel(value string) Label {
	return Label{strValue: value, isInt: false}
}
