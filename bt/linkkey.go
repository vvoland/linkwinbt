package bt

import (
	"fmt"
	"strings"
)

// LinkKey represents a Bluetooth link key
// It's a 32 length uppercase hex string
// 9EF14F8CB54D8B01048F4A8F4A8F4A8F
type LinkKey string

// ParseLinkKey parses a LinkKey from a string, possible formats:
// 32 len hex string: 9EF14F8CB54D8B01048F4A8F4A8F4A8F
// Windows reg hex: hex:c5,cc,96,ec,48,ee,88,8f,04,a8,63,34,4c,c6,a7,2d
func ParseLinkKey(str string) (LinkKey, error) {
	str = strings.TrimSpace(str)
	str = strings.TrimPrefix(str, "hex:")
	str = strings.ReplaceAll(str, ",", "")
	str = strings.ToUpper(str)
	if len(str) != 32 {
		return "", fmt.Errorf("invalid link key format: %s", str)
	}
	return LinkKey(str), nil
}

func (l LinkKey) String() string {
	return string(l)
}
