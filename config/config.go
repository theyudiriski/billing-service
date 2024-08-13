package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func RequireEnv(key string) string {
	value, hasValue := os.LookupEnv(key)
	if !hasValue {
		panic(fmt.Errorf("%s is a required env variable", key))
	}

	return strings.TrimSpace(value)
}

func RequireEnvToDuration(key string) time.Duration {
	value, err := time.ParseDuration(RequireEnv(key))
	if err != nil {
		panic(fmt.Errorf("%s should be a time.Duration", key))
	}

	return value
}

func RequireEnvToInt(key string) int {
	value, err := strconv.Atoi(RequireEnv(key))
	if err != nil {
		panic(fmt.Errorf("%s should be a number", key))
	}
	return value
}

func RequireEnvToBool(key string) bool {
	value := RequireEnv(key)
	switch value {
	case "true":
		return true
	case "false":
		return false
	default:
		panic(fmt.Errorf("%s should be true or false", key))
	}
}

func ReadPrivateKey(encodedKey string) (*rsa.PrivateKey, error) {
	pemBytes, err := base64.StdEncoding.DecodeString(encodedKey)
	if err != nil {
		return nil, fmt.Errorf("ReadPrivateKey error: %w", err)
	}

	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, fmt.Errorf("ReadPrivateKey error: decoding failed")
	}

	parseResult, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("ReadPrivateKey error: %w", err)
	}

	privateKey, ok := parseResult.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("ReadPrivateKey error: parsed key is not rsa private key")
	}

	return privateKey, nil
}

func ReadPublicKey(encodedKey string) (*rsa.PublicKey, error) {
	pemBytes, err := base64.StdEncoding.DecodeString(encodedKey)
	if err != nil {
		return nil, fmt.Errorf("ReadPublicKey error: %w", err)
	}
	block, _ := pem.Decode(pemBytes)
	if block.Type != "PUBLIC KEY" {
		return nil, errors.New("ReadPublicKey error: Found wrong key type")
	}

	parseResult, _ := x509.ParsePKIXPublicKey(block.Bytes)
	publicKey, ok := parseResult.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("ReadPublicKey error: parsed key is not rsa public key")
	}

	return publicKey, nil
}

func OptionalEnv(key, defaultVal string) string {
	value, hasValue := os.LookupEnv(key)
	if !hasValue || value == "" {
		return defaultVal
	}

	return strings.TrimSpace(value)
}

func OptionalEnvToDuration(key string, defaultVal time.Duration) time.Duration {
	strVal, hasValue := os.LookupEnv(key)
	if !hasValue || strVal == "" {
		return defaultVal
	}

	value, err := time.ParseDuration(strVal)
	if err != nil {
		panic(fmt.Errorf("%s should be a time.Duration", key))
	}

	return value
}

func OptionalEnvToInt(key string, defaultVal int) int {
	strVal, hasValue := os.LookupEnv(key)
	if !hasValue || strVal == "" {
		return defaultVal
	}

	value, err := strconv.Atoi(strVal)
	if err != nil {
		panic(fmt.Errorf("%s should be a number", key))
	}

	return value
}

func OptionalEnvToBool(key string, defaultVal bool) bool {
	value, hasValue := os.LookupEnv(key)
	if !hasValue {
		return defaultVal
	}

	switch value {
	case "true":
		return true
	case "false":
		return false
	default:
		panic(fmt.Errorf("%s should be true or false", key))
	}
}

func OptionalEnvToTime(key string, defaultVal time.Time) time.Time {
	strVal, hasValue := os.LookupEnv(key)
	if !hasValue || strVal == "" {
		return defaultVal
	}

	value, err := time.Parse(time.RFC3339, strVal)
	if err != nil {
		panic(fmt.Errorf("%s should be a valid timestamp, ex 2024-05-17T14:00:00+08:00", key))
	}

	return value
}

func RequireEnvToFloat64(key string) float64 {
	value, err := strconv.ParseFloat(RequireEnv(key), 64)
	if err != nil {
		panic(fmt.Errorf("%s should be a float64", key))
	}

	return value
}

func OptionalEnvToFloat(key string, defaultVal float64) float64 {
	strVal, hasValue := os.LookupEnv(key)
	if !hasValue || strVal == "" {
		return defaultVal
	}

	value, err := strconv.ParseFloat(strVal, 64)
	if err != nil {
		panic(fmt.Errorf("%s should be a float64", key))
	}

	return value
}

func RequireEnvToStringSlice(key string) []string {
	strVal, hasValue := os.LookupEnv(key)
	if !hasValue || strVal == "" {
		panic(fmt.Errorf("%s is a required env variable", key))
	}

	return strings.Split(strVal, ",")
}

func OptionalEnvToStringSlice(key string, defaultVal []string) []string {
	strVal, hasValue := os.LookupEnv(key)
	if !hasValue || strVal == "" {
		return defaultVal
	}

	return strings.Split(strVal, ",")
}

func OptionalEnvToIntSlice(key string, defaultVal []int) []int {
	strVal, hasValue := os.LookupEnv(key)
	if !hasValue || strVal == "" {
		return defaultVal
	}

	stringSlice := strings.Split(strVal, ",")

	var value []int

	for _, s := range stringSlice {
		v, err := strconv.Atoi(s)
		if err != nil {
			panic(fmt.Errorf("%v should be a number", v))
		}
		value = append(value, v)
	}

	return value
}
