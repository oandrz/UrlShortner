package data

import (
	"slices"
	"strings"
)

const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

/*
*

digit           next number
3844 % 62 = 0    3844 / 62 = 62

	  62 % 62 = 0      62 / 62 = 1
	   1 % 62 = 1       1 / 62 = 0   <- stop
	digits collected: 0, 0, 1
*/
func EncodeBase62(number uint64) string {
	if number == 0 {
		return "0"
	}

	digitArr := make([]uint64, 0, 10)
	nextNumber := number
	for nextNumber > 0 {
		digit := nextNumber % 62
		digitArr = append(digitArr, digit)
		nextNumber = nextNumber / 62
	}

	var sb strings.Builder
	for _, digit := range slices.Backward(digitArr) {
		sb.WriteByte(alphabet[digit])
	}

	return sb.String()
}

/*
**
g8 -> 1000
100 -> 3844

g 8

g - 16
8 - 8

g - 16 = 16 * (62^1) = 992
8 - 8 = 8 * (62^0) = 8
992 + 8 = 1000

1 - 1 * (62^2) = 3844
0 - 0 * (62^1) = 0
0 - 0 * (62^0) = 0
*/
func decodeBase62(digits string) uint64 {
	var number uint64 = 0
	exponential := len(digits) - 1

	for _, digit := range digits {
		trueNumber := pow(62, uint64(exponential))
		number += uint64(strings.Index(alphabet, string(digit))) * trueNumber
		exponential--
	}

	return number
}

func pow(base, exp uint64) uint64 {
	var result uint64 = 1

	for exp > 0 {
		result = result * base
		exp--
	}

	return result
}
