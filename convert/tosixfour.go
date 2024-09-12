package convert

import (
	"errors"
	"strconv"
	"strings"
)

const baseChars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-_"
const LIMIT = 1073741822

func ToSixFour(row int64) (string, error) {

	if row > LIMIT || row < 1 {
		return "", errors.New("row of " + strconv.FormatInt(row, 10) + " not in allowed range")
	}

	var result strings.Builder
	for row > 0 {
		result.WriteString(string(baseChars[row%64]))
		row /= 64
	}
	return reverse(result.String()), nil

}

func reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
