package hw02unpackstring

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	runePattern      = `(?P<Rune>\P{N})(?P<Repeats>\d?)`
	ErrInvalidString = errors.New("invalid string")
)

func Unpack(inputString string) (string, error) {
	regEx, err := regexp.Compile(fmt.Sprintf("^(%s)*$", runePattern))
	if err != nil {
		return "", err
	}
	matched := regEx.MatchString(inputString)
	if !matched {
		return "", ErrInvalidString
	}
	regEx, _ = regexp.Compile(runePattern)
	matches := regEx.FindAllStringSubmatch(inputString, -1)
	var outputBuilder strings.Builder
	for _, match := range matches {
		if match[2] == "" {
			match[2] = "1"
		}
		repeats, _ := strconv.Atoi(match[2])
		for i := 0; i < repeats; i++ {
			outputBuilder.WriteString(match[1])
		}
	}
	return outputBuilder.String(), nil
}
