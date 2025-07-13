package collector

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
)

func GenerateV5(str1 string, str2 string, space string) string {
	uuid, _ := v5(fmt.Sprintf("%s/%s", str1, str2), space)
	return uuid
}

// Converts a namespace UUID string to bytes
func uuidToBytes(ns string) ([]byte, error) {
	return hex.DecodeString(ns)
}

// Generates a UUID v5 from a namespace and value
func v5(value, nspace string) (string, error) {
	// Convert namespace UUID to bytes
	nsBytes, err := uuidToBytes(nspace)
	if err != nil {
		return "", fmt.Errorf("invalid namespace UUID: %v", err)
	}

	// SHA-1 hashing
	hasher := sha1.New()
	hasher.Write(nsBytes)
	hasher.Write([]byte(value))
	hash := hasher.Sum(nil)

	// Set version (5) and variant (RFC 4122)
	hash[6] = (hash[6] & 0x0f) | 0x50 // Set version to 5
	hash[8] = (hash[8] & 0x3f) | 0x80 // Set variant to RFC 4122

	// Format UUID string
	uuid := fmt.Sprintf(UUIDFormat,
		hash[0], hash[1], hash[2], hash[3],
		hash[4], hash[5],
		hash[6], hash[7],
		hash[8], hash[9],
		hash[10], hash[11], hash[12], hash[13], hash[14], hash[15],
	)

	return uuid, nil
}
