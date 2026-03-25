package domain

import (
	"regexp"
	"strings"
)

var inputWordPattern = regexp.MustCompile(`[\pL\pN]+`)

func Classify(input string) InputType {
	for _, token := range inputWordPattern.FindAllString(strings.ToLower(input), -1) {
		switch token {
		case "inventory", "character", "quest", "save", "quit", "help":
			return MetaAction
		}
	}

	return GameAction
}
