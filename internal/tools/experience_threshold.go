package tools

import "git.subcult.tv/subculture-collective/edda/internal/progression"

type experienceThresholdFunc func(level int) int

func defaultExperienceThreshold(level int) int {
	return progression.NextLevelThreshold(level)
}
