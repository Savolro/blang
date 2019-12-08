package generate

// Rule defines any rule defined in BNF
type Rule struct {
	Name    string
	Options []RuleOption
}

// RuleOption defines one of rule options
type RuleOption struct {
	IsConstant bool
	Value      []string
}

// IsRecursive returns true if rule contains itself as its option
// Note: This does not consider descendant recursion
func (r Rule) IsRecursive() bool {
	for _, option := range r.Options {
		if Contains(option.Value, r.Name) {
			return true
		}
	}
	return false
}
