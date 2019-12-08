package main

import (
	"fmt"
	"github.com/Savolro/blang/lexer/pkg/states"
	"github.com/Savolro/blang/lexer/pkg/states/generate"
	"github.com/Savolro/blang/lexer/pkg/states/generate/bnf"
	"github.com/olekukonko/tablewriter"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strconv"
)

func main() {
	bytes, err := ioutil.ReadFile("/home/savolro/dev/uni/blang/blang_lexems.bnf")
	if err != nil {
		fmt.Printf("Error on reading BNF input: %s", err.Error())
		return
	}
	rules, err := bnf.Parse(bytes)
	if err != nil {
		fmt.Printf("Error on parsing BNF input: %s", err.Error())
		return
	}
	generator := generate.NewGenerator(rules)
	table := generator.Generate()
	// tg := bnf.NewTableGenerator(rules)
	// table := tg.GetTable()
	// printTableTransitions(table)
	printTable(table, true)
	//doNothing(table)
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

func doNothing(table states.Table) {

}

func printTableTransitions(table states.Table) {
	for _, fromState := range table {
		for char, toID := range fromState.Transitions {
			log.Println("From:", fromState.Name, "Path:", string(char), "To:", table[toID].Name, fmt.Sprintf("(%s)", table[toID].Lexeme))
		}
	}
}

func printTable(table states.Table, stringify bool) {
	var columns []string
	for _, s := range table {
		for c := range s.Transitions {
			columns = append(columns, string(c))
		}
	}
	columns = distinctStr(columns)
	sort.Strings(columns)
	columns = append([]string{"STATE", "LEXEME"}, columns...)
	t := tablewriter.NewWriter(os.Stdout)
	t.SetAutoFormatHeaders(false)
	t.SetHeader(columns)

	for i := 1; i <= len(table); i++ {
		state := table[i]
		stateStr := state.Name
		if !stringify {
			stateStr = strconv.Itoa(i)
		}
		data := []string{stateStr, state.Lexeme}
		for i := 2; i < len(columns); i++ {
			byteID := []byte(columns[i])[0]
			id := state.Transitions[byteID]

			var toStr string
			if stringify {
				toStr = table[id].Name
			} else {
				toStr = strconv.Itoa(id)
			}
			data = append(data, toStr)
		}
		t.Append(data)
	}
	t.Render()
}

func contains(list []string, s string) bool {
	for _, str := range list {
		if str == s {
			return true
		}
	}
	return false
}

func distinctStr(list []string) (res []string) {
	for _, str := range list {
		if !contains(res, str) {
			res = append(res, str)
		}
	}
	return res
}
