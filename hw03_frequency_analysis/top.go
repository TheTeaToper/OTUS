package hw03frequencyanalysis

import (
	"cmp"
	"fmt"
	"regexp"
	"slices"
	"strings"
)

type word struct {
	Value   string
	Counter int
}

var regEx = regexp.MustCompile(`[\pL-]+`)

func Top10(inputText string) []string {
	matches := regEx.FindAllStringSubmatch(inputText, -1)
	wordsMap := map[string]int{}
	for _, match := range matches {
		wordsMap[strings.ToLower(match[0])]++
	}
	delete(wordsMap, "-")
	words := make([]word, 0, len(wordsMap))
	for k, v := range wordsMap {
		words = append(words, word{k, v})
	}
	slices.SortFunc(words, func(a, b word) int {
		if sortResult := cmp.Compare(b.Counter, a.Counter); sortResult != 0 {
			return sortResult
		}
		return cmp.Compare(a.Value, b.Value)
	})
	result := make([]string, min(10, len(words)))
	for i := 0; i < len(result); i++ {
		result[i] = words[i].Value
		fmt.Println(words[i].Value, words[i].Counter)
	}
	return result
}
