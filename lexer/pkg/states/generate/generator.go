package generate

import (
	"github.com/Savolro/blang/lexer/pkg/states"
	"log"
	"strconv"
	"strings"
)

// TableGenerator is responsible for generating lexeme table from given rules
type Generator struct {
	rules map[string]Rule
}

// NewGenerator is a default constructor for TableGenerator
func NewGenerator(rules map[string]Rule) *Generator {
	return &Generator{
		rules: rules,
	}
}

// Generate generates a state table from predefined rules
func (g *Generator) Generate() states.Table {
	table := states.Table{}
	// Initiate a start state, which is default for any lexer using this generator
	table[1] = states.NewState("START", true, "")
	a := g.getRuleMap()
	g.generateStates("DIGIT", table, a)
	return table
}

func (g *Generator) generateStates(lexeme string, table states.Table, rulesMap map[string]*ruleGroup) {
	g.next(table, 1, 0, []string{}, rulesMap[lexeme])
}

func (g *Generator) next(table states.Table, currentIndex int, nextIndex int, path []string, currentGroup *ruleGroup) {
	if Contains(path, currentGroup.Key) {
		return
	}
	if nextIndex == 0 {
		nextIndex = len(table) + 1
	}
	if table[nextIndex].Name == "" {
		table[nextIndex] = states.NewState(currentGroup.Key+strconv.Itoa(currentIndex), false, "TEST")
	}
	switch currentGroup.Type {
	case SINGLE:
		table[currentIndex].Transitions[currentGroup.Value] = nextIndex // current state
	case ORDER:
	case CHOICE:
		// same previous state
	}

	for i, group := range currentGroup.Groups {
		if currentGroup.Type == ORDER {
			if i > 0 {
				currentIndex = nextIndex
				nextIndex = currentIndex + 1
			}
		}
		g.next(table, currentIndex, nextIndex, append(path, currentGroup.Key), group)
	}
}

// ruleGroup defines group of rules concatenated with CHOICE or ORDER
type ruleGroup struct {
	Key    string
	Type   groupType
	Groups []*ruleGroup
	Value  byte
}

type groupType int

const (
	CHOICE groupType = iota
	ORDER
	SINGLE
)

func (g *Generator) getRuleGroup(rules []string, currentMap map[string]*ruleGroup) *ruleGroup {
	key := strings.Join(rules, ":")
	keyGroup := currentMap[key]
	if keyGroup != nil {
		return keyGroup
	}
	group := &ruleGroup{Key: key, Type: ORDER}

	for _, ruleName := range rules {
		rule := g.rules[ruleName]
		rGroup := &ruleGroup{Key: ruleName, Type: CHOICE}
		for _, option := range rule.Options {
			optionKey := strings.Join(option.Value, ":")
			optionGroup := &ruleGroup{Key: optionKey, Type: ORDER}
			if option.IsConstant {
				// Constants will always have only 1 value
				val := option.Value[0]
				for _, c := range val {
					charGroup := &ruleGroup{
						Type:  SINGLE,
						Value: byte(c),
					}
					if len(val) == 1 {
						optionGroup = charGroup
					} else {
						optionGroup.Groups = append(optionGroup.Groups, charGroup)
					}
				}
			} else {
				currentMap[key] = &ruleGroup{}
				optionGroup = g.getRuleGroup(option.Value, currentMap)
				currentMap[key] = optionGroup
			}
			if len(rule.Options) == 1 {
				rGroup = optionGroup
			} else {
				rGroup.Groups = append(group.Groups, optionGroup)
				if rule.Name == "DIGIT" {
					log.Println(optionGroup.Type, string(optionGroup.Value))
				}
			}
		}
		if len(rules) > 1 {
			group.Groups = append(group.Groups, rGroup)
		} else {
			group = rGroup
		}
	}
	currentMap[key] = group
	return group
}

// getPossibleCharacters returns list of possible next characters after specified one
func (g *Generator) getRuleMap() map[string]*ruleGroup {
	ruleMap := map[string]*ruleGroup{}
	for rule := range g.rules {
		g.getRuleGroup([]string{rule}, ruleMap)
	}
	return ruleMap
}

// getLexemesList returns a list of lexemes
func (g *Generator) getLexemesList() (lexemes []string) {
	return []string{
		"IDENT",
		"LIT_FLOAT",
		"KW_IF",
	}
}
