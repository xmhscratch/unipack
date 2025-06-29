package uni

import (
	"fmt"
	"testing"
)

func TestKK(t *testing.T) {
	mode := 0070
	fmt.Println(mode & 0111)
}
