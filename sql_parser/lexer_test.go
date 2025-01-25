package sql_parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLexer(t *testing.T) {
	tests := []struct {
		input    string
		expected []token
	}{
		{
			input: "SELECT name, age FROM users;",
			expected: []token{
				{Type: tokenType_RESERVED, Literal: "SELECT"},
				{Type: tokenType_IDENTIFIER, Literal: "name"},
				{Type: tokenType_SYMBOL, Literal: ","},
				{Type: tokenType_IDENTIFIER, Literal: "age"},
				{Type: tokenType_RESERVED, Literal: "FROM"},
				{Type: tokenType_IDENTIFIER, Literal: "users"},
				{Type: tokenType_SYMBOL, Literal: ";"},
				{Type: tokenType_EOF, Literal: ""},
			},
		},
		{
			input: "SELECT name AS name_alt FROM users;",
			expected: []token{
				{Type: tokenType_RESERVED, Literal: "SELECT"},
				{Type: tokenType_IDENTIFIER, Literal: "name"},
				{Type: tokenType_RESERVED, Literal: "AS"},
				{Type: tokenType_IDENTIFIER, Literal: "name_alt"},
				{Type: tokenType_RESERVED, Literal: "FROM"},
				{Type: tokenType_IDENTIFIER, Literal: "users"},
				{Type: tokenType_SYMBOL, Literal: ";"},
				{Type: tokenType_EOF, Literal: ""},
			},
		},
		{
			input: "SELECT name `name_alt` FROM users;",
			expected: []token{
				{Type: tokenType_RESERVED, Literal: "SELECT"},
				{Type: tokenType_IDENTIFIER, Literal: "name"},
				{Type: tokenType_IDENTIFIER, Literal: "name_alt"},
				{Type: tokenType_RESERVED, Literal: "FROM"},
				{Type: tokenType_IDENTIFIER, Literal: "users"},
				{Type: tokenType_SYMBOL, Literal: ";"},
				{Type: tokenType_EOF, Literal: ""},
			},
		},
		{
			input: "SELECT name, age FROM `users` WHERE age > 18;",
			expected: []token{
				{Type: tokenType_RESERVED, Literal: "SELECT"},
				{Type: tokenType_IDENTIFIER, Literal: "name"},
				{Type: tokenType_SYMBOL, Literal: ","},
				{Type: tokenType_IDENTIFIER, Literal: "age"},
				{Type: tokenType_RESERVED, Literal: "FROM"},
				{Type: tokenType_IDENTIFIER, Literal: "users"},
				{Type: tokenType_RESERVED, Literal: "WHERE"},
				{Type: tokenType_IDENTIFIER, Literal: "age"},
				{Type: tokenType_SYMBOL, Literal: ">"},
				{Type: tokenType_NUMBER, Literal: "18"},
				{Type: tokenType_SYMBOL, Literal: ";"},
				{Type: tokenType_EOF, Literal: ""},
			},
		},
		{
			input: "SELECT * FROM `users` WHERE `id` = ? AND `name` = 'Alice';",
			expected: []token{
				{Type: tokenType_RESERVED, Literal: "SELECT"},
				{Type: tokenType_SYMBOL, Literal: "*"},
				{Type: tokenType_RESERVED, Literal: "FROM"},
				{Type: tokenType_IDENTIFIER, Literal: "users"},
				{Type: tokenType_RESERVED, Literal: "WHERE"},
				{Type: tokenType_IDENTIFIER, Literal: "id"},
				{Type: tokenType_SYMBOL, Literal: "="},
				{Type: tokenType_SYMBOL, Literal: "?"},
				{Type: tokenType_RESERVED, Literal: "AND"},
				{Type: tokenType_IDENTIFIER, Literal: "name"},
				{Type: tokenType_SYMBOL, Literal: "="},
				{Type: tokenType_STRING, Literal: "Alice"},
				{Type: tokenType_SYMBOL, Literal: ";"},
				{Type: tokenType_EOF, Literal: ""},
			},
		},
		{
			input: "SELECT * FROM `users` WHERE `id` = ? AND `name` = 'Alice' ORDER BY `id` ASC LIMIT 10 OFFSET 0;",
			expected: []token{
				{Type: tokenType_RESERVED, Literal: "SELECT"},
				{Type: tokenType_SYMBOL, Literal: "*"},
				{Type: tokenType_RESERVED, Literal: "FROM"},
				{Type: tokenType_IDENTIFIER, Literal: "users"},
				{Type: tokenType_RESERVED, Literal: "WHERE"},
				{Type: tokenType_IDENTIFIER, Literal: "id"},
				{Type: tokenType_SYMBOL, Literal: "="},
				{Type: tokenType_SYMBOL, Literal: "?"},
				{Type: tokenType_RESERVED, Literal: "AND"},
				{Type: tokenType_IDENTIFIER, Literal: "name"},
				{Type: tokenType_SYMBOL, Literal: "="},
				{Type: tokenType_STRING, Literal: "Alice"},
				{Type: tokenType_RESERVED, Literal: "ORDER BY"},
				{Type: tokenType_IDENTIFIER, Literal: "id"},
				{Type: tokenType_RESERVED, Literal: "ASC"},
				{Type: tokenType_RESERVED, Literal: "LIMIT"},
				{Type: tokenType_NUMBER, Literal: "10"},
				{Type: tokenType_RESERVED, Literal: "OFFSET"},
				{Type: tokenType_NUMBER, Literal: "0"},
				{Type: tokenType_SYMBOL, Literal: ";"},
				{Type: tokenType_EOF, Literal: ""},
			},
		},
		{
			input: "UPDATE `users` SET `name` = 'Bob', `age` = 20 WHERE `id` = ?;",
			expected: []token{
				{Type: tokenType_RESERVED, Literal: "UPDATE"},
				{Type: tokenType_IDENTIFIER, Literal: "users"},
				{Type: tokenType_RESERVED, Literal: "SET"},
				{Type: tokenType_IDENTIFIER, Literal: "name"},
				{Type: tokenType_SYMBOL, Literal: "="},
				{Type: tokenType_STRING, Literal: "Bob"},
				{Type: tokenType_SYMBOL, Literal: ","},
				{Type: tokenType_IDENTIFIER, Literal: "age"},
				{Type: tokenType_SYMBOL, Literal: "="},
				{Type: tokenType_NUMBER, Literal: "20"},
				{Type: tokenType_RESERVED, Literal: "WHERE"},
				{Type: tokenType_IDENTIFIER, Literal: "id"},
				{Type: tokenType_SYMBOL, Literal: "="},
				{Type: tokenType_SYMBOL, Literal: "?"},
				{Type: tokenType_SYMBOL, Literal: ";"},
				{Type: tokenType_EOF, Literal: ""},
			},
		},
		{
			input: "DELETE FROM `users` WHERE `id` = ?;",
			expected: []token{
				{Type: tokenType_RESERVED, Literal: "DELETE"},
				{Type: tokenType_RESERVED, Literal: "FROM"},
				{Type: tokenType_IDENTIFIER, Literal: "users"},
				{Type: tokenType_RESERVED, Literal: "WHERE"},
				{Type: tokenType_IDENTIFIER, Literal: "id"},
				{Type: tokenType_SYMBOL, Literal: "="},
				{Type: tokenType_SYMBOL, Literal: "?"},
				{Type: tokenType_SYMBOL, Literal: ";"},
				{Type: tokenType_EOF, Literal: ""},
			},
		},
		{
			input: "INSERT INTO `users` (`name`, `age`) VALUES ('Cathy', 30);",
			expected: []token{
				{Type: tokenType_RESERVED, Literal: "INSERT"},
				{Type: tokenType_RESERVED, Literal: "INTO"},
				{Type: tokenType_IDENTIFIER, Literal: "users"},
				{Type: tokenType_SYMBOL, Literal: "("},
				{Type: tokenType_IDENTIFIER, Literal: "name"},
				{Type: tokenType_SYMBOL, Literal: ","},
				{Type: tokenType_IDENTIFIER, Literal: "age"},
				{Type: tokenType_SYMBOL, Literal: ")"},
				{Type: tokenType_RESERVED, Literal: "VALUES"},
				{Type: tokenType_SYMBOL, Literal: "("},
				{Type: tokenType_STRING, Literal: "Cathy"},
				{Type: tokenType_SYMBOL, Literal: ","},
				{Type: tokenType_NUMBER, Literal: "30"},
				{Type: tokenType_SYMBOL, Literal: ")"},
				{Type: tokenType_SYMBOL, Literal: ";"},
				{Type: tokenType_EOF, Literal: ""},
			},
		},
		{
			input: "SELECT COUNT(*) AS `count` FROM `comments` WHERE `post_id` = ?",
			expected: []token{
				{Type: tokenType_RESERVED, Literal: "SELECT"},
				{Type: tokenType_RESERVED, Literal: "COUNT"},
				{Type: tokenType_SYMBOL, Literal: "("},
				{Type: tokenType_SYMBOL, Literal: "*"},
				{Type: tokenType_SYMBOL, Literal: ")"},
				{Type: tokenType_RESERVED, Literal: "AS"},
				{Type: tokenType_IDENTIFIER, Literal: "count"},
				{Type: tokenType_RESERVED, Literal: "FROM"},
				{Type: tokenType_IDENTIFIER, Literal: "comments"},
				{Type: tokenType_RESERVED, Literal: "WHERE"},
				{Type: tokenType_IDENTIFIER, Literal: "post_id"},
				{Type: tokenType_SYMBOL, Literal: "="},
				{Type: tokenType_SYMBOL, Literal: "?"},
				{Type: tokenType_EOF, Literal: ""},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			lexer := NewLexer(test.input)
			tokens := []token{}
			for {
				tok := lexer.NextToken()
				tokens = append(tokens, tok)
				if tok.Type == tokenType_EOF {
					break
				}
			}
			assert.Equal(t, test.expected, tokens)
		})
	}
}
