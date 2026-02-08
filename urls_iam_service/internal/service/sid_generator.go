package service

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

type SIDGenerator struct {
	length int
}

func NewSIDGenerator(length int) *SIDGenerator {
	if length < 16 {
		length = 16
	}
	return &SIDGenerator{length: length}
}

func (s *SIDGenerator) Generate() (string, error) {
	b := make([]byte, s.length)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (g *SIDGenerator) GenerateAnonymousProviderID() (string, error) {
	b := make([]byte, 16)

	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return fmt.Sprintf("anon_%s", base64.RawURLEncoding.EncodeToString(b)), nil
}
