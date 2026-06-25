package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	jwtSecret  = "llm-relay-fixed-secret-please-change-in-production"
	jwtIssuer  = "llm-relay"
	jwtAlgorithm = "HS256"
)

type Claims struct {
	Sub string `json:"sub"`
	Iss string `json:"iss"`
	Iat int64  `json:"iat"`
	Exp int64  `json:"exp"`
}

type jwtHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

func base64URLEncode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

func base64URLDecode(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}

func computeSignature(signingInput string) string {
	mac := hmac.New(sha256.New, []byte(jwtSecret))
	mac.Write([]byte(signingInput))
	return base64URLEncode(mac.Sum(nil))
}

func GenerateToken(subject string, ttl time.Duration) (string, error) {
	now := time.Now()

	headerBytes, err := json.Marshal(jwtHeader{Alg: jwtAlgorithm, Typ: "JWT"})
	if err != nil {
		return "", fmt.Errorf("failed to encode header: %w", err)
	}
	headerEncoded := base64URLEncode(headerBytes)

	claims := Claims{
		Sub: subject,
		Iss: jwtIssuer,
		Iat: now.Unix(),
		Exp: now.Add(ttl).Unix(),
	}
	claimsBytes, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("failed to encode claims: %w", err)
	}
	claimsEncoded := base64URLEncode(claimsBytes)

	signingInput := headerEncoded + "." + claimsEncoded
	return signingInput + "." + computeSignature(signingInput), nil
}

func ValidateToken(token string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid token format")
	}

	signingInput := parts[0] + "." + parts[1]
	expectedSig := computeSignature(signingInput)
	if !hmac.Equal([]byte(parts[2]), []byte(expectedSig)) {
		return nil, errors.New("invalid signature")
	}

	headerBytes, err := base64URLDecode(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid header: %w", err)
	}
	var header jwtHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, fmt.Errorf("invalid header: %w", err)
	}
	if header.Alg != jwtAlgorithm {
		return nil, fmt.Errorf("unexpected algorithm: %s", header.Alg)
	}

	claimsBytes, err := base64URLDecode(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid claims: %w", err)
	}
	var claims Claims
	if err := json.Unmarshal(claimsBytes, &claims); err != nil {
		return nil, fmt.Errorf("invalid claims: %w", err)
	}

	if claims.Iss != jwtIssuer {
		return nil, errors.New("invalid issuer")
	}
	if time.Now().Unix() >= claims.Exp {
		return nil, errors.New("token expired")
	}

	return &claims, nil
}
