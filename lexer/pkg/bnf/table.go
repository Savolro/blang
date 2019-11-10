package bnf

import (
	"fmt"
	"sort"
	"strings"
)

type TransitionRange struct {
	Min byte
	Max byte
}

type TableGenerator struct {
	rules            map[string]Rule
	checkedRules     map[string]int
	checkedEntries   map[string][]byte
	checkedRecursive map[string]bool
}

func NewTableGenerator(rules map[string]Rule) *TableGenerator {
	return &TableGenerator{
		rules:            rules,
		checkedRules:     map[string]int{},
		checkedEntries:   map[string][]byte{},
		checkedRecursive: map[string]bool{},
	}
}

func (g *TableGenerator) getLexemesList() (lexemes []string) {
	for name := range g.rules {
		if g.isLexeme(name) {
			lexemes = append(lexemes, name)
		}
	}
	return lexemes
}

type State struct {
	Name        string
	ConsistsOf  []string
	MustIndex   int
	Transitions map[byte]int
}

func NewState(name string, consistsOf []string, mustIndex int) State {
	return State{
		Name:        name,
		ConsistsOf:  consistsOf,
		MustIndex:   mustIndex,
		Transitions: map[byte]int{},
	}
}

func (g *TableGenerator) GetTable() map[int]State {
	g.checkedRules = map[string]int{}
	table := map[int]State{}

	table[0] = NewState("START", nil, -1)

	totalChars, maxIndex := g.getTotalChars()

	stateIndex := 1
	for name := range totalChars {
		table[stateIndex] = NewState(name, []string{name}, -1)
		stateIndex++
	}

	for i := 0; i < maxIndex; i++ {
		groups := g.getOverlappingGroups(totalChars, i)
		for _, group := range groups {
			sort.Strings(group)
			name := fmt.Sprintf("%s.%d", strings.Join(group, "."), i)
			table[stateIndex] = NewState(name, group, i)
			stateIndex++
		}
	}

	return table
}

func (g *TableGenerator) getStateEntryChars(state State, index int) []byte {
	if state.MustIndex != -1 && state.MustIndex != index || len(state.ConsistsOf) == 0 {
		return nil
	}
	if len(state.ConsistsOf) == 1 {
		_, chars := g.getEntryChars(state.Name, index)
		return chars
	}
	_, currentChars := g.getEntryChars(state.ConsistsOf[0], index)
	for _, lexeme := range state.ConsistsOf {
		_, lexemeChars := g.getEntryChars(lexeme, index)
		for i, c := range currentChars {
			if !containsBytes(lexemeChars, c) {
				currentChars[i] = currentChars[len(currentChars)-1]
				currentChars[len(currentChars)-1] = 0
				currentChars = currentChars[:len(currentChars)-1]
			}
		}
	}
	return nil
}

func (g *TableGenerator) isRecursive(name string, checkedNames []string) bool {
	if contains(checkedNames, name) {
		g.checkedRecursive[name] = true
		return true
	}
	for _, option := range g.rules[name].Options {
		if !option.IsConstant {
			for _, val := range option.Value {
				if g.isRecursive(val, append(checkedNames, name)) {
					g.checkedRecursive[name] = true
					return true
				}
			}
		}
	}
	g.checkedRecursive[name] = false
	return false
}

func (g *TableGenerator) getOverlappingGroups(totalChars map[string][][]byte, index int) (groups [][]string) {
	var names []string
	for name := range totalChars {
		names = append(names, name)
	}
	combinations := All(names)
	for _, c := range combinations {
		if g.combinationMatches(totalChars, c, index) {
			groups = append(groups, c)
		}
	}
	sort.Slice(groups, func(i int, j int) bool {
		return len(groups[i]) < len(groups[j])
	})
	removedIndexes := map[int]bool{}
	for i, gr := range groups {
		for j := i + 1; j < len(groups); j++ {
			if containsAll(groups[j], gr) {
				removedIndexes[i] = true
			}
		}
	}
	var splitGroups [][]string
	for i, gr := range groups {
		if removedIndexes[i] != true && len(gr) >= 2 {
			splitGroups = append(splitGroups, gr)
		}
	}
	return splitGroups
}

func (g *TableGenerator) combinationMatches(totalChars map[string][][]byte, combination []string, index int) bool {
	for _, name1 := range combination {
		for _, name2 := range combination {
			if name1 == name2 {
				continue
			}
			if !matchesUntilPosition(totalChars[name1], g.checkedRecursive[name1], totalChars[name2], g.checkedRecursive[name2], index) {
				return false
			}
		}
	}
	return true
}

// All returns all combinations for a given string array.
// This is essentially a powerset of the given set except that the empty set is disregarded.
func All(set []string) (subsets [][]string) {
	length := uint(len(set))

	// Go through all possible combinations of objects
	// from 1 (only first object in subset) to 2^length (all objects in subset)
	for subsetBits := 1; subsetBits < (1 << length); subsetBits++ {
		var subset []string

		for object := uint(0); object < length; object++ {
			// checks if object is contained in subset
			// by checking if bit 'object' is set in subsetBits
			if (subsetBits>>object)&1 == 1 {
				// add object to subset
				subset = append(subset, set[object])
			}
		}
		// add subset to subsets
		subsets = append(subsets, subset)
	}
	return subsets
}

func (g *TableGenerator) getTotalChars() (map[string][][]byte, int) {
	// get lexemes list
	lexemes := g.getLexemesList()
	totalChars := map[string][][]byte{}
	// Assuming that all lexemes begin with START state
	maxIndex := 0
	for _, lexeme := range lexemes {
		isRecursive, entryChars := g.getEntryChars(lexeme, 0)
		list := totalChars[lexeme]
		for i := 1; !isRecursive && len(entryChars) > 0; i++ {
			list = append(list, entryChars)
			isRecursive, entryChars = g.getEntryChars(lexeme, i)
		}
		if len(list) > maxIndex {
			maxIndex = len(list)
		}
		totalChars[lexeme] = list
	}
	return totalChars, maxIndex
}

func (g *TableGenerator) getUsedChars(totalChars map[string][][]byte, maxIndex int) map[int]map[byte][]string {
	usedChars := map[int]map[byte][]string{}
	for i := 0; i < maxIndex; i++ {
		usedChars[i] = map[byte][]string{}
		for lexeme, list := range totalChars {
			// Select the rightest symbol if this is a long keyword or recursive and append all i'th characters
			// to usedChars
			isRecursive := g.checkedRecursive[lexeme]
			if len(list) > i || isRecursive {
				index := i
				if len(list) <= i && isRecursive {
					index = len(list) - 1
				}
				l := list[index]
				for _, char := range l {
					usedChars[i][char] = append(usedChars[i][char], lexeme)
				}
			}
		}
	}
	return usedChars
}

func matchesUntilPosition(bytes1 [][]byte, recursive1 bool, bytes2 [][]byte, recursive2 bool, position int) bool {
	for i := 0; i <= position; i++ {
		if !hasSameCharInPosition(bytes1, recursive1, bytes2, recursive2, i) {
			return false
		}
	}
	return true
}

func hasSameCharInPosition(bytes1 [][]byte, recursive1 bool, bytes2 [][]byte, recursive2 bool, position int) bool {
	if !recursive1 && position >= len(bytes1) {
		return false
	}
	if !recursive2 && position >= len(bytes2) {
		return false
	}
	var l1 []byte
	var l2 []byte
	if len(bytes1) > position {
		l1 = bytes1[position]
	} else {
		l1 = bytes1[len(bytes1)-1]
	}
	if len(bytes2) > position {
		l2 = bytes2[position]
	} else {
		l2 = bytes2[len(bytes2)-1]
	}
	for _, c := range l1 {
		if containsBytes(l2, c) {
			return true
		}
	}
	return false
}

func (g *TableGenerator) getEntryChars(name string, index int) (isRecursive bool, chars []byte) {
	// Take the leftest characters
	for _, option := range g.rules[name].Options {
		if !option.IsConstant {
			if contains(option.Value, name) {
				isRecursive = true
			}
			if len(option.Value) > index {
				var nextChars []byte
				isRecursive, nextChars = g.getEntryChars(option.Value[index], 0)
				chars = append(chars, nextChars...)
			}
		} else {
			for _, val := range option.Value {
				if len(val) > index {
					chars = append(chars, val[index])
				}
			}
		}
	}
	return isRecursive, distinct(chars)
}

func (g *TableGenerator) isLexeme(ruleName string) bool {
	if g.checkedRules[ruleName] != 0 {
		return g.checkedRules[ruleName] == 1
	}
	g.checkedRules[ruleName] = 2
	usedCount := 0
	rule := g.rules[ruleName]
	if len(rule.Options) == 1 && len(rule.Options[0].Value) == 1 && rule.Options[0].IsConstant {
		g.checkedRules[ruleName] = 1
		return true
	}

	for _, r := range g.rules {
		if r.Name != ruleName {
			for _, option := range r.Options {
				if !option.IsConstant && contains(option.Value, rule.Name) {
					usedCount++
				}
			}
		}
	}
	if usedCount > 1 {
		childLexeme := false
		for _, option := range rule.Options {
			if !option.IsConstant {
				for _, val := range option.Value {
					if g.isLexeme(val) {
						childLexeme = true
						break
					}
				}
				if childLexeme {
					break
				}
			}
		}
		if childLexeme || !g.isRecursive(ruleName, []string{}) {
			g.checkedRules[ruleName] = 2
			return false
		} else {
			g.checkedRules[ruleName] = 1
			return true
		}
	}
	g.checkedRules[ruleName] = 2
	return false
}

func containsAll(listIn []string, list []string) bool {
	for _, s := range list {
		if !contains(listIn, s) {
			return false
		}
	}
	return true
}

func contains(list []string, s string) bool {
	for _, str := range list {
		if str == s {
			return true
		}
	}
	return false
}

func containsBytes(list []byte, b byte) bool {
	for _, bt := range list {
		if bt == b {
			return true
		}
	}
	return false
}

func remove(list []string, s string) []string {
	for i, str := range list {
		if str == s {
			list[i] = list[len(list)-1]
			list[len(list)-1] = ""
			list = list[:len(list)-1]
			return list
		}
	}
	return list
}

func removeByte(bytes []byte, b byte) []byte {
	for i, bt := range bytes {
		if bt == b {
			bytes[i] = bytes[len(bytes)-1]
			bytes[len(bytes)-1] = 0
			bytes = bytes[:len(bytes)-1]
			return bytes
		}
	}
	return bytes
}

func distinctList(list [][]string) (res [][]string) {
	for _, l1 := range list {
		skip := false
		for _, l2 := range res {
			containsCount := 0
			for _, str := range l1 {
				if contains(l2, str) {
					containsCount++
				}
			}
			if containsCount == len(l1) {
				skip = true
			}
		}
		if !skip {
			res = append(res, l1)
		}
	}
	return res
}

func distinctStr(list []string) (res []string) {
	for _, str := range list {
		if !contains(res, str) {
			res = append(res, str)
		}
	}
	return res
}

func distinct(bytes []byte) (res []byte) {
	for _, b := range bytes {
		if !containsBytes(res, b) {
			res = append(res, b)
		}
	}
	return res
}
