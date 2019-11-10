package main

import (
	"fmt"
	"io/ioutil"

	"github.com/Savolro/tm-lexer/pkg/bnf"
)

func main() {
	bytes, err := ioutil.ReadFile("/home/savolro/dev/tm/blang_lexems.bnf")
	if err != nil {
		fmt.Printf("Error on reading BNF input: %s", err.Error())
		return
	}
	rules, err := bnf.Parse(bytes)
	if err != nil {
		fmt.Printf("Error on parsing BNF input: %s", err.Error())
		return
	}
	tg := bnf.NewTableGenerator(rules)
	table := tg.GetTable()
	// for fromState, v := range table {
	// 	for char, toState := range v.Transitions {
	// 		fmt.Println("From:", kwNames[fromState], "Path:", string(char), "To:", kwNames[toState])
	// 	}
	// }
	fmt.Println(table)
	// filename := os.Args[1]
	// bytes, err := ioutil.ReadFile(filename)
	// if err != nil {
	// 	fmt.Printf("Error on opening file: %s", err.Error())
	// 	return
	// }

	// lexer := lexicon.NewLexer()
	// tokens, err := lexer.GetTokens(bytes)
	// if err != nil {
	// 	fmt.Printf("%s: Lexer error: %s\n", filename, err.Error())
	// 	return
	// }

	// t := tablewriter.NewWriter(os.Stdout)
	// t.SetHeader([]string{"ID", "LN", "COL", "TYPE", "VALUE"})
	// for i, token := range tokens {
	// 	t.Append([]string{strconv.Itoa(i), strconv.Itoa(token.Line), strconv.Itoa(token.Column), token.Type, token.Value})
	// }
	// t.Render()
}
