package convert

import (
	"errors"
	"strconv"
	"strings"
)

const baseChars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-_"

func ToSixFour(row int) (string, error) {

	if row > 2147483647 || row < 1 {
		return "", errors.New("row of " + strconv.Itoa(row) + " not in allowed range")
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
