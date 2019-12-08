package bnf

import (
	"errors"
	"fmt"
	"github.com/Savolro/blang/lexer/pkg/states/generate"
	"strings"
)

// Parse parses slice of BNF content lines into BNF rule map
func Parse(content []byte) (rules map[string]generate.Rule, err error) {
	rules = map[string]generate.Rule{}
	lines := strings.Split(string(content), "\n")
	for _, line := range lines[:len(lines)-1] {
		line = strings.Replace(line, " ", "", -1)
		line = strings.Replace(line, "\t", "", -1)
		// Ignore comments
		if strings.HasPrefix(line, ";") || line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		parts := strings.Split(line, "::=")
		if len(parts) != 2 {
			return nil, errors.New("each line should be in format: <RULE> ::= options")
		}
		ruleName := strings.TrimSuffix(strings.TrimPrefix(parts[0], "<"), ">")
		if rules[ruleName].Name != "" {
			return nil, fmt.Errorf("rule %s is defined multiple times", ruleName)
		}

		rule := generate.Rule{
			Name: ruleName,
		}

		partsOfParts := strings.Split(parts[1], "|")
		for _, part := range partsOfParts {
			var value []string
			var isConstant bool
			if hasPrefixSuffix(part, "<", ">") {
				value = strings.Split(trimPrefixSuffix(part, "<", ">"), "><")
				isConstant = false
			} else if hasPrefixSuffix(part, "\"", "\"") {
				value = []string{trimPrefixSuffix(part, "\"", "\"")}
				isConstant = true
			} else {
				return nil, errors.New("options should either point to rules or be constants")
			}
			rule.Options = append(rule.Options, generate.RuleOption{
				IsConstant: isConstant,
				Value:      value,
			})
		}
		rules[ruleName] = rule
	}
	for _, rule := range rules {
		for _, option := range rule.Options {
			if !option.IsConstant {
				for _, val := range option.Value {
					if rules[val].Name == "" {
						return nil, fmt.Errorf("rule %s is not defined", val)
					}
				}
			}
		}
	}
	return rules, nil
}

func hasPrefixSuffix(s string, prefix string, suffix string) bool {
	return strings.HasPrefix(s, prefix) && strings.HasSuffix(s, suffix)
}

func trimPrefixSuffix(s string, prefix string, suffix string) string {
	return strings.TrimSuffix(strings.TrimPrefix(s, prefix), suffix)
}
