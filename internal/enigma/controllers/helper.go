package controllers

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"

	"google.golang.org/grpc/metadata"
)

// Character to use, exclude 0,1,I,O
var (
	charMap = []string{
		"U", "V", "W", "X", "Y", "Z",
		"A", "B", "C", "D", "E", "F", "G", "H", "J",
		"2", "3", "4", "5", "6", "7", "8", "9",
		"K", "L", "M", "N", "P", "Q", "R", "S", "T",
	}
)

const (
	codePrefix = "MO" //Manabie order
	randLen    = 3
	randMin    = 101
	randMax    = 999
)

func getRandValue(seed uint, randMin, randMax, randLen int) int {
	localRand := rand.New(rand.NewSource(int64(seed)))
	rawValue := fmt.Sprintf("%d", localRand.Intn((randMax-randMin)+1)+randMin)
	randValue, _ := strconv.Atoi(rawValue[:randLen])
	return randValue
}

func dec2n(val int64) string {
	base := len(charMap)
	var result string
	for {
		index := val % int64(base)
		result = charMap[index] + result
		if val = int64(math.Floor(float64(val) / float64(base))); val == 0 {
			break
		}
	}
	return result
}

func reserveString(s string) (result string) {
	for _, v := range s {
		result = string(v) + result
	}

	return
}

func sliceIndex(length int, f func(i int) bool) int {
	for i := 0; i < length; i++ {
		if f(i) {
			return i
		}
	}

	return -1
}

func n2dec(val string) (string, error) {
	base := len(charMap)

	var decode int64 = 0
	val = reserveString(val)
	for digit := 0; digit < len(val); digit++ {
		char := string(val[digit])
		if !isContain(charMap, char) {
			return "", errors.New("invalid input")
		}

		index := int64(sliceIndex(len(charMap), func(i int) bool { return charMap[i] == char }))
		decode += index * int64(math.Pow(float64(base), float64(digit)))
	}
	return fmt.Sprintf("%d", decode), nil
}

func isContain(arr []string, value string) bool {
	for _, item := range arr {
		if item == value {
			return true
		}
	}
	return false
}

func EncodeNumber2String(n uint) string {
	randVal := getRandValue(n, randMin, randMax, randLen)

	idTemp := fmt.Sprintf("%03d%09d", randVal, n)
	id, _ := strconv.ParseInt(idTemp, 10, 64)
	return codePrefix + dec2n(id)
}

func DecodeString2Number(encoded string) (uint, error) {
	encoded = strings.Replace(encoded, codePrefix, "", -1)
	decoded, err := n2dec(encoded)
	if err != nil {
		return 0, err
	}
	if len(decoded) < randLen+1 {
		return 0, errors.New("invalid encoded")
	}
	id, _ := strconv.ParseInt(decoded[randLen:], 10, 64)
	decodedRandVal, _ := strconv.Atoi(decoded[:randLen])

	randVal := getRandValue(uint(id), randMin, randMax, randLen)
	if randVal != decodedRandVal {
		return 0, errors.New("invalid rand value")
	}

	return uint(id), nil
}

func contextWithValidVersion(ctx context.Context, clientName, clientVersion string) context.Context {
	return metadata.AppendToOutgoingContext(ctx,
		"pkg",
		clientName,
		"version",
		clientVersion)
}
