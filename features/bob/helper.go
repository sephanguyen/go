package bob

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
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

func (s *suite) EncodeOrderID2String(n uint) string {
	return codePrefix + EncodeNumber2String(n)
}

func EncodeNumber2String(n uint) string {
	randVal := getRandValue(n, randMin, randMax, randLen)

	idTemp := fmt.Sprintf("%03d%09d", randVal, n)
	id, _ := strconv.ParseInt(idTemp, 10, 64)
	return dec2n(id)
}
