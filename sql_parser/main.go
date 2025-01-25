package sql_parser

import "fmt"

// <sql> := <select-stmt> | <update-stmt> | <delete-stmt> | <insert-stmt>
// <select-stmt> := SELECT <select-values> FROM <table> [WHERE <conditions>] [ORDER BY <orders>] [LIMIT <limit>] [OFFSET <offset>] ;
// <select-values> := <select-value> [, <select-values>]
// <select-value> := (<column> | COUNT(<select-value>) | SUM(<select-value>) | AVG(<select-value>) | MIN(<select-value>) | MAX(<select-value>) | *) [<select-alias>]
// <select-alias> := [AS] <column>
// <update-stmt> := UPDATE <table> SET <update-sets> [WHERE <conditions>] [ORDER BY <orders>] [LIMIT <limit>] [OFFSET <offset>] ;
// <update-sets> := <column> = <value> [, <update-sets>]
// <delete-stmt> := DELETE FROM <table> [WHERE <conditions>] [ORDER BY <orders>] [LIMIT <limit>] [OFFSET <offset>] ;
// <insert-stmt> := INSERT INTO <table> (<columns>) VALUES (<values>) ;
// <conditions> := <condition> [AND <conditions>]
// <condition> := <column> <operator> <value>
// <orders> := <order> [, <orders>]
// <order> := <column> [ASC | DESC]
// <limit> := <number> | ?
// <offset> := <number> | ?
// <columns> := <column> [, <columns>]
// <column> := <identifier> | `<identifier>`
// <table> := <identifier> | `<identifier>`
// <operator> := = | != | < | > | <= | >=
// <values> := <value> [, <values>]
// <value> := <string> | <number> | ? | (<values>)

func ParseSQL(sql string) (SQLNode, error) {
	lexer := NewLexer(sql)
	tokens := []token{}
	for {
		token := lexer.NextToken()
		tokens = append(tokens, token)
		if token.Type == tokenType_EOF {
			break
		}
	}
	parser := NewParser(tokens)
	node, err := parser.Parse()
	if err != nil {
		return nil, fmt.Errorf("failed to parse: %w", err)
	}
	return node, nil
}
