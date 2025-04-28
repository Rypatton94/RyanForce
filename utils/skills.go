package utils

import (
	"encoding/json"
	"strings"
)

// ParseSkills takes a JSON array string and returns a slice of skills.
func ParseSkills(skillsJSON string) ([]string, error) {
	var skills []string
	err := json.Unmarshal([]byte(skillsJSON), &skills)
	if err != nil {
		return nil, err
	}
	return skills, nil
}

// MatchScore returns how many skills in neededSkills are present in techSkills.
func MatchScore(techSkills, neededSkills []string) int {
	score := 0
	skillSet := make(map[string]struct{})
	for _, skill := range techSkills {
		skillSet[strings.ToLower(skill)] = struct{}{}
	}

	for _, need := range neededSkills {
		if _, ok := skillSet[strings.ToLower(need)]; ok {
			score++
		}
	}

	return score
}
