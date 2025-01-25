package domains

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var formatted = `queries:
  - query: SELECT * FROM livecomments WHERE livestream_id = ? ORDER BY created_at DESC
    type: select
    table: livecomments
    cache: true
    targets:
      - id
      - user_id
      - livestream_id
      - comment
      - tip
      - created_at
    conditions:
      - column: livestream_id
        operator: eq
        placeholder:
          index: 0
    orders:
      - column: created_at
        order: desc
  - query: SELECT r.emoji_name FROM users u INNER JOIN livestreams l ON l.user_id = u.id INNER JOIN reactions r ON r.livestream_id = l.id WHERE u.name = ? GROUP BY emoji_name ORDER BY COUNT(*) DESC, emoji_name DESC LIMIT ?
    type: select
    cache: false
  - query: SELECT COUNT(*) FROM livestream_viewers_history WHERE livestream_id = ?
    type: select
    table: livestream_viewers_history
    cache: true
    targets:
      - COUNT()
    conditions:
      - column: livestream_id
        operator: eq
        placeholder:
          index: 0
  - query: DELETE FROM livecomments WHERE id = ? AND livestream_id = ? AND (SELECT COUNT(*) FROM (SELECT ? AS text) AS texts INNER JOIN (SELECT CONCAT('%', ?, '%') AS pattern) AS patterns ON texts.text LIKE patterns.pattern) >= 1;
    type: delete
    table: livecomments
    conditions:
      - column: id
        operator: eq
        placeholder:
          index: 0
      - column: livestream_id
        operator: eq
        placeholder:
          index: 1
  - query: DELETE FROM livestream_viewers_history WHERE user_id = ? AND livestream_id = ?
    type: delete
    table: livestream_viewers_history
    conditions:
      - column: user_id
        operator: eq
        placeholder:
          index: 0
      - column: livestream_id
        operator: eq
        placeholder:
          index: 1
  - query: UPDATE settings SET value = ? WHERE name = ?
    type: update
    table: settings
    targets:
      - column: value
        placeholder:
          index: 0
    conditions:
      - column: name
        operator: eq
        placeholder:
          index: 0
          extra: true
  - query: INSERT INTO livecomments (user_id, livestream_id, comment, tip, created_at) VALUES (?)
    type: insert
    table: livecomments
    columns:
      - user_id
      - livestream_id
      - comment
      - tip
      - created_at
`

var parsed = &CachePlan{
	Queries: []*CachePlanQuery{
		{
			CachePlanQueryBase: &CachePlanQueryBase{
				Type:  CachePlanQueryType_SELECT,
				Query: "SELECT * FROM livecomments WHERE livestream_id = ? ORDER BY created_at DESC",
			},
			Select: &CachePlanSelectQuery{
				Table:   "livecomments",
				Cache:   true,
				Targets: []string{"id", "user_id", "livestream_id", "comment", "tip", "created_at"},
				Conditions: []CachePlanCondition{
					{Column: "livestream_id", Operator: CachePlanOperator_EQ, Placeholder: CachePlanPlaceholder{Index: 0}},
				},
				Orders: []CachePlanOrder{
					{Column: "created_at", Order: "desc"},
				},
			},
		},
		{
			CachePlanQueryBase: &CachePlanQueryBase{
				Type:  CachePlanQueryType_SELECT,
				Query: "SELECT r.emoji_name FROM users u INNER JOIN livestreams l ON l.user_id = u.id INNER JOIN reactions r ON r.livestream_id = l.id WHERE u.name = ? GROUP BY emoji_name ORDER BY COUNT(*) DESC, emoji_name DESC LIMIT ?",
			},
			Select: &CachePlanSelectQuery{
				Cache: false,
			},
		},
		{
			CachePlanQueryBase: &CachePlanQueryBase{
				Query: "SELECT COUNT(*) FROM livestream_viewers_history WHERE livestream_id = ?",
				Type:  CachePlanQueryType_SELECT,
			},
			Select: &CachePlanSelectQuery{
				Table:   "livestream_viewers_history",
				Cache:   true,
				Targets: []string{"COUNT()"},
				Conditions: []CachePlanCondition{
					{Column: "livestream_id", Operator: CachePlanOperator_EQ, Placeholder: CachePlanPlaceholder{Index: 0}},
				},
			},
		},
		{
			CachePlanQueryBase: &CachePlanQueryBase{
				Query: "DELETE FROM livecomments WHERE id = ? AND livestream_id = ? AND (SELECT COUNT(*) FROM (SELECT ? AS text) AS texts INNER JOIN (SELECT CONCAT('%', ?, '%') AS pattern) AS patterns ON texts.text LIKE patterns.pattern) >= 1;",
				Type:  CachePlanQueryType_DELETE,
			},
			Delete: &CachePlanDeleteQuery{
				Table: "livecomments",
				Conditions: []CachePlanCondition{
					{Column: "id", Operator: CachePlanOperator_EQ, Placeholder: CachePlanPlaceholder{Index: 0}},
					{Column: "livestream_id", Operator: CachePlanOperator_EQ, Placeholder: CachePlanPlaceholder{Index: 1}},
				},
			},
		},
		{
			CachePlanQueryBase: &CachePlanQueryBase{
				Query: "DELETE FROM livestream_viewers_history WHERE user_id = ? AND livestream_id = ?",
				Type:  CachePlanQueryType_DELETE,
			},
			Delete: &CachePlanDeleteQuery{
				Table: "livestream_viewers_history",
				Conditions: []CachePlanCondition{
					{Column: "user_id", Operator: CachePlanOperator_EQ, Placeholder: CachePlanPlaceholder{Index: 0}},
					{Column: "livestream_id", Operator: CachePlanOperator_EQ, Placeholder: CachePlanPlaceholder{Index: 1}},
				},
			},
		},
		{
			CachePlanQueryBase: &CachePlanQueryBase{
				Query: "UPDATE settings SET value = ? WHERE name = ?",
				Type:  CachePlanQueryType_UPDATE,
			},
			Update: &CachePlanUpdateQuery{
				Table: "settings",
				Targets: []CachePlanUpdateTarget{
					{Column: "value", Placeholder: CachePlanPlaceholder{Index: 0}},
				},
				Conditions: []CachePlanCondition{
					{Column: "name", Operator: CachePlanOperator_EQ, Placeholder: CachePlanPlaceholder{Index: 0, Extra: true}},
				},
			},
		},
		{
			CachePlanQueryBase: &CachePlanQueryBase{
				Query: "INSERT INTO livecomments (user_id, livestream_id, comment, tip, created_at) VALUES (?)",
				Type:  CachePlanQueryType_INSERT,
			},
			Insert: &CachePlanInsertQuery{
				Table:   "livecomments",
				Columns: []string{"user_id", "livestream_id", "comment", "tip", "created_at"},
			},
		},
	},
}

func TestLoadCachePlan(t *testing.T) {
	reader := strings.NewReader(formatted)

	plan, err := LoadCachePlan(reader)
	assert.NoError(t, err)
	assert.Equal(t, parsed, plan)
}

func TestSaveCachePlan(t *testing.T) {
	writer := &strings.Builder{}
	err := SaveCachePlan(writer, parsed)
	assert.NoError(t, err)
	assert.Equal(t, formatted, writer.String())
}
