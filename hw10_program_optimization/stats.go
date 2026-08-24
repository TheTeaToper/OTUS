package hw10programoptimization

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"
)

type User struct {
	Email string
}

type DomainStat map[string]int

func GetDomainStat(r io.Reader, domain string) (DomainStat, error) {
	regExp, err := regexp.Compile(`(?i)\.` + domain + `$`)
	if err != nil {
		return nil, fmt.Errorf("regex compilation error: %w", err)
	}

	domainStat := make(DomainStat)
	jsonDecoder := json.NewDecoder(r)

	for {
		var user User
		err := jsonDecoder.Decode(&user)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("json decoding error: %w", err)
		}
		if regExp.MatchString(user.Email) {
			emailParts := strings.Split(user.Email, "@")
			if len(emailParts) == 2 {
				extractedDomain := strings.ToLower(emailParts[1])
				domainStat[extractedDomain]++
			}
		}
	}

	return domainStat, nil
}
