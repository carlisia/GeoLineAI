// Package ids is the application's identifier type and its one wire format.
package ids

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
)

// ID is an integer the application assigns. On the wire it is a decimal string,
// which clients treat as opaque.
type ID int64

// ReadingIDRule says what a reading ID on the wire must be, after the field it concerns.
var ReadingIDRule = Rule("reading ID")

// Rule says what an ID of the named kind must be on the wire, after the field it
// concerns.
func Rule(kind string) string { return "must be a " + kind + `, a decimal string such as "1".` }

var canonical = regexp.MustCompile(`^[1-9][0-9]*$`)

// Parse accepts only the canonical form: a decimal string with no sign, no
// leading zeros and no zero, within int64.
func Parse(s string) (ID, error) {
	if !canonical.MatchString(s) {
		return 0, fmt.Errorf("%q is not an ID", s)
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%q is not an ID", s)
	}
	return ID(n), nil
}

func (id ID) String() string { return strconv.FormatInt(int64(id), 10) }

// MarshalJSON writes the ID as a JSON string.
func (id ID) MarshalJSON() ([]byte, error) { return json.Marshal(id.String()) }
