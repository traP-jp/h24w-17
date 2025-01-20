package domains

type CachePlan struct {
	Queries []CachePlanQuery `yaml:"queries"`
}

type CachePlanQuery struct {
	*CachePlanQueryBase
	Select *CachePlanSelectQuery
	Update *CachePlanUpdateQuery
	Delete *CachePlanDeleteQuery
	Insert *CachePlanInsertQuery
}

type CachePlanQueryBase struct {
	Type  CachePlanQueryType `yaml:"type"`
	Query string             `yaml:"query"`
}

type CachePlanQueryType string

const (
	CachePlanQueryType_SELECT CachePlanQueryType = "select"
	CachePlanQueryType_UPDATE CachePlanQueryType = "update"
	CachePlanQueryType_DELETE CachePlanQueryType = "delete"
	CachePlanQueryType_INSERT CachePlanQueryType = "insert"
)

type CachePlanCondition struct {
	Column string `yaml:"column"`
	Value  string `yaml:"value,omitempty"`
}

type CachePlanOrder struct {
	Column string             `yaml:"column"`
	Order  CachePlanOrderEnum `yaml:"order"`
}

type CachePlanOrderEnum string

const (
	CachePlanOrder_ASC  CachePlanOrderEnum = "asc"
	CachePlanOrder_DESC CachePlanOrderEnum = "desc"
)

type CachePlanSelectQuery struct {
	CachePlanQueryBase
	Table      string               `yaml:"table"`
	Cache      bool                 `yaml:"cache"`
	Targets    []string             `yaml:"targets"`
	Conditions []CachePlanCondition `yaml:"conditions,omitempty"`
	Orders     []CachePlanOrder     `yaml:"orders,omitempty"`
}

type CachePlanUpdateQuery struct {
	CachePlanQueryBase
	Table      string               `yaml:"table"`
	Targets    []string             `yaml:"targets"`
	Conditions []CachePlanCondition `yaml:"conditions,omitempty"`
}

type CachePlanDeleteQuery struct {
	CachePlanQueryBase
	Table      string               `yaml:"table"`
	Conditions []CachePlanCondition `yaml:"conditions,omitempty"`
}

type CachePlanInsertQuery struct {
	Table string `yaml:"table"`
	CachePlanQueryBase
}
