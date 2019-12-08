package bnf

/*
import (
	"fmt"
	"log"
	"sort"
	"strings"
)

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

func (t *Table) GetStateID(name string) int {
	for i, s := range *t {
		if s.Name == name {
			return i
		}
	}
	return -1
}

func (g *TableGenerator) GetTable() map[int]State {
	g.checkedRules = map[string]int{}
	table := Table{}

	table[1] = NewState("START", nil, -1, true, "")

	totalChars, maxIndex := g.getTotalChars()

	stateIndex := 2

	for i := 0; i < maxIndex; i++ {
		groups := g.getOverlappingGroups(totalChars, i)

		for _, group := range groups {
			lenChars := len(totalChars[group[0]])
			if len(group) == 1 && i > lenChars-1 {
				continue
			}

			lexeme := ""
			name := ""
			for _, s := range group {
				if !g.checkedRecursive[s] && i == len(totalChars[s])-1 {
					name = s
					lexeme = s
					break
				}
			}
			if name == "" {
				for _, s := range group {
					if g.checkedRecursive[s] {
						lexeme = s
						if len(group) == 1 && i == len(totalChars[s])-1 {
							name = s
						}
						break
					}
				}
			}

			if name == "" {
				name = fmt.Sprintf("%s.%d", strings.Join(group, "."), i)
			}

			mustIndex := i
			if len(group) == 1 && i == lenChars-1 && g.checkedRecursive[group[0]] {
				mustIndex = -1
			}

			table[stateIndex] = NewState(name, group, mustIndex, false, lexeme)
			stateIndex++
		}
	}

	table.updateTransitions(totalChars, g.checkedRecursive)
	return table
}

func (t *Table) updateTransitions(totalChars map[string][][]byte, checkedRecursive map[string]bool) {
	maxIndex := 0
	for _, s := range *t {
		if s.MustIndex > maxIndex {
			maxIndex = s.MustIndex
		}
	}

	order := t.getTransitionOrder()
	for i := maxIndex + 1; i >= 0; i-- {
		for _, id1 := range order {
			s1 := (*t)[id1]
			if s1.MustIndex != i-1 && s1.MustIndex != -1 {
				continue
			}
			var usedChars []byte

			for _, id2 := range order {
				s2 := (*t)[id2]
				if s2.MustIndex != i || !containsAll(s1.ConsistsOf, s2.ConsistsOf) && id1 != 1 ||
					s1.MustIndex == -1 && id1 != 1 || id1 == 1 && i > 0 {
					continue
				}
				if s1.Name == "START" {
					log.Println(i, s2.Name)
				}

				state0Name := s2.ConsistsOf[0]
				commonChars := getCharsByIndex(state0Name, totalChars, i, checkedRecursive[state0Name])
				for _, name := range s2.ConsistsOf {
					commonChars = getCommonChars(commonChars, getCharsByIndex(name, totalChars, i, checkedRecursive[name]))
				}
				for _, c := range commonChars {
					(*t)[id1].Transitions[c] = id2
					usedChars = append(usedChars, c)
				}
			}
			usedChars = distinct(usedChars)
			for _, s := range s1.ConsistsOf {
				id := t.GetStateID(s)
				state := (*t)[(*t).GetStateID(s)]
				recursive := state.MustIndex == -1
				if !recursive && s1.Name == state.Name {
					continue
				}
				for _, c := range getCharsByIndex(s, totalChars, i, checkedRecursive[s]) {
					if !containsBytes(usedChars, c) && (*t)[id1].Transitions[c] == 0 {
						(*t)[id1].Transitions[c] = id
					}
				}
			}
		}
	}
}

func (t *Table) getTransitionOrder() []int {
	var states []State
	var order []int
	for _, st := range *t {
		states = append(states, st)
	}
	sort.Slice(states, func(i int, j int) bool {
		stateI := states[i]
		stateJ := states[j]
		if len(stateI.ConsistsOf) != len(stateJ.ConsistsOf) {
			return len(stateI.ConsistsOf) < len(stateJ.ConsistsOf)
		}
		return stateI.MustIndex < stateJ.MustIndex
	})
	for _, state := range states {
		order = append(order, t.GetStateID(state.Name))
	}
	return order
}

func getIndex(initIndex int, length int) int {
	if initIndex < 0 || initIndex >= length {
		return length - 1
	}
	return initIndex
}

func getCommonChars(s1 []byte, s2 []byte) []byte {
	var chars []byte
	for _, c := range s1 {
		if containsBytes(s2, c) && !containsBytes(chars, c) {
			chars = append(chars, c)
		}
	}
	for _, c := range s2 {
		if containsBytes(s1, c) && !containsBytes(chars, c) {
			chars = append(chars, c)
		}
	}
	return chars
}

func (g *TableGenerator) getStateEntryChars(state State, index int) []byte {
	if state.MustIndex != -1 && state.MustIndex != index || len(state.ConsistsOf) == 0 {
		return nil
	}
	if len(state.ConsistsOf) == 1 {
		chars := g.getEntryChars(state.Name, index)
		return chars
	}
	currentChars := g.getEntryChars(state.ConsistsOf[0], index)
	for _, lexeme := range state.ConsistsOf {
		lexemeChars := g.getEntryChars(lexeme, index)
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
			// remove only if all chars are used
			if containsAll(groups[j], gr) {
				for _, s := range gr {
					for _, s1 := range groups[j] {
						sChars := getCharsByIndex(s, totalChars, index, g.checkedRecursive[s])
						s1Chars := getCharsByIndex(s1, totalChars, index, g.checkedRecursive[s1])
						if !containsAllBytes(sChars, s1Chars) || s != s1 && containsAllBytes(s1Chars, sChars) {
							removedIndexes[i] = true
						}
					}
				}
			}
		}
	}
	var splitGroups [][]string
	for i, gr := range groups {
		if !removedIndexes[i] {
			splitGroups = append(splitGroups, gr)
		}
	}
	return splitGroups
}

func getCharsByIndex(name string, totalChars map[string][][]byte, index int, recursive bool) []byte {
	listChars := totalChars[name]
	if recursive {
		index = getIndex(index, len(listChars))
	}
	if index >= len(listChars) {
		return nil
	}
	return listChars[index]
}

func (g *TableGenerator) combinationMatches(totalChars map[string][][]byte, combination []string, index int) bool {
	for _, name1 := range combination {
		for _, name2 := range combination {
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
		entryChars := g.getEntryChars(lexeme, 0)
		list := totalChars[lexeme]
		for i := 1; len(entryChars) > 0; i++ {
			if g.checkedRecursive[lexeme] && len(list) > 0 && string(list[i-2]) == string(entryChars) {
				break
			}
			list = append(list, entryChars)
			entryChars = g.getEntryChars(lexeme, i)
		}
		if len(list) > maxIndex {
			maxIndex = len(list)
		}
		totalChars[lexeme] = list
	}
	return totalChars, maxIndex
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

func (g *TableGenerator) getEntryChars(name string, index int) (chars []byte) {
	// Take the leftest characters
	for _, option := range g.rules[name].Options {
		if !option.IsConstant {
			if len(option.Value) > index {
				nextChars := g.getEntryChars(option.Value[index], 0)
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
	return distinct(chars)
}

func (g *TableGenerator) isLexeme(ruleName string) bool {
	if g.checkedRules[ruleName] != 0 {
		return g.checkedRules[ruleName] == 1
	}
	g.checkedRules[ruleName] = 2
	usedCount := 0
	rule := g.rules[ruleName]
	if strings.HasPrefix(rule.Name, "LIT_") {
		g.checkedRules[ruleName] = 1
		return true
	}
	if len(rule.Options) == 1 && len(rule.Options[0].Value) == 1 && rule.Options[0].IsConstant && !strings.HasSuffix(rule.Name, "_C") {
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

func containsAllBytes(listIn []byte, list []byte) bool {
	for _, s := range list {
		if !containsBytes(listIn, s) {
			return false
		}
	}
	return true
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
*/
