package extractor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractQueryFromFile(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		root     string
		expected []*ExtractedQuery
	}{
		{
			name: "extract query from file",
			path: "testdata/extractor1.go",
			root: "testdata",
			expected: []*ExtractedQuery{
				{file: "extractor1.go", pos: 32, content: "SELECT id, name FROM users"},
				{file: "extractor1.go", pos: 44, content: "INSERT INTO users (name) VALUES (?)"},
				{file: "extractor1.go", pos: 59, content: "SELECT id, name FROM users WHERE id = ?"},
				{file: "extractor1.go", pos: 72, content: "UPDATE users SET name = ? WHERE id = ?"},
				{file: "extractor1.go", pos: 79, content: "DELETE FROM users WHERE id = ?"},
				{file: "extractor1.go", pos: 89, content: "SELECT id, user_id, title, body FROM posts"},
				{file: "extractor1.go", pos: 103, content: "INSERT INTO posts (user_id, title, body) VALUES (?, ?, ?)"},
				{file: "extractor1.go", pos: 118, content: "SELECT id, user_id, title, body FROM posts WHERE id = ?"},
				{file: "extractor1.go", pos: 133, content: "UPDATE posts SET user_id = ?, title = ?, body = ? WHERE id = ?"},
				{file: "extractor1.go", pos: 140, content: "DELETE FROM posts WHERE id = ?"},
				{file: "extractor1.go", pos: 150, content: "SELECT id, post_id, body FROM comments"},
				{file: "extractor1.go", pos: 163, content: "INSERT INTO comments (post_id, body) VALUES (?, ?)"},
				{file: "extractor1.go", pos: 178, content: "SELECT id, post_id, body FROM comments WHERE id = ?"},
				{file: "extractor1.go", pos: 191, content: "UPDATE comments SET body = ? WHERE id = ?"},
				{file: "extractor1.go", pos: 198, content: "DELETE FROM comments WHERE id = ?"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := ExtractQueryFromFile(test.path, test.root)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, actual)
		})
	}
}
