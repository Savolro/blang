package lexicon

// Token defines a single token found in a program
type Token struct {
	Type   string
	Value  string
	Line   int
	Column int
}
