package uni

import (
	"fmt"
	"strings"
)

// StringMapFlag implements flag.Value for a map[string]string
type StringMapFlag map[string]string

// Set parses the string argument into the map
func (sm StringMapFlag) Set(value string) error {
	pairs := strings.Split(value, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			sm[parts[0]] = parts[1]
		} else {
			return fmt.Errorf("invalid map format: %s", pair)
		}
	}
	return nil
}

// String returns the string representation of the map
func (sm StringMapFlag) String() string {
	var s []string
	for k, v := range sm {
		s = append(s, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(s, ",")
}
