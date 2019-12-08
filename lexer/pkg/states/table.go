package states

// Table defines all lexeme states and transitions
type Table map[int]State

// State defines a single state and it's possible transitions for a lexer
type State struct {
	Name        string
	IsConstant  bool
	Lexeme      string
	Transitions map[byte]int
}

// NewState is a default constructor for State
func NewState(name string, isConstant bool, lexeme string) State {
	return State{
		Name:        name,
		IsConstant:  isConstant,
		Lexeme:      lexeme,
		Transitions: map[byte]int{},
	}
}
