package convert

import (
	"errors"
	"math"
	"strings"
)

func FromSixFour(str string) (int64, error) {
	if len(str) > 6 {
		return 0, errors.New("too large string to be actual int64")
	}
	if len(str) == 0 {
		return 0, errors.New("blank str not allowed")
	}
	var result int64
	for i, char := range str {
		power := len(str) - i - 1
		index := strings.IndexRune(baseChars, char)
		if index == -1 {
			return 0, errors.New("character used in id parameter not allowed")
		}
		result += int64(index) * int64(math.Pow(float64(64), float64(power)))
	}
	if result > 2147483647 {
		return 0, errors.New("end result higher than allowed max")
	}
	return result, nil
}
