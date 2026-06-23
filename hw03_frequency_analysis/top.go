package hw03frequencyanalysis

import (
	"cmp"
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/dlclark/regexp2"
	// github.com/dlclark/regexp2/syntax требуется для активации расширенных таблиц Unicode (Emoji_Presentation).
	_ "github.com/dlclark/regexp2/syntax"
)

type word struct {
	Value   string
	Counter int
}

var regEx = regexp2.MustCompile(`[\pL\p{Emoji_Presentation}-]+`, 0)

func Top10(inputText string) []string {
	matches, err := FindAllMatches(regEx, inputText)
	if err != nil {
		log.Fatal(err)
	}
	wordsMap := map[string]int{}
	for _, match := range matches {
		wordsMap[strings.ToLower(match)]++
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

func FindAllMatches(re *regexp2.Regexp, text string) ([]string, error) {
	var results []string

	// Ищем первое совпадение
	m, err := re.FindStringMatch(text)
	if err != nil {
		return nil, err
	}

	// Итерируемся по всем остальным совпадениям
	for m != nil {
		results = append(results, m.String()) // Добавляем найденное слово/эмодзи в срез

		m, err = re.FindNextMatch(m) // Переходим к следующему совпадению
		if err != nil {
			return nil, err
		}
	}

	return results, nil
}
