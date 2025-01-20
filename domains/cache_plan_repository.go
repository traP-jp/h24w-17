package domains

import (
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

func (c *CachePlanQuery) UnmarshalYAML(value *yaml.Node) error {
	var base CachePlanQueryBase
	if err := value.Decode(&base); err != nil {
		return fmt.Errorf("failed to decode cache plan query base: %w", err)
	}
	c.CachePlanQueryBase = &base

	switch base.Type {
	case CachePlanQueryType_SELECT:
		var query CachePlanSelectQuery
		if err := value.Decode(&query); err != nil {
			return fmt.Errorf("failed to decode cache plan select query: %w", err)
		}
		c.Select = &query
	case CachePlanQueryType_UPDATE:
		var query CachePlanUpdateQuery
		if err := value.Decode(&query); err != nil {
			return fmt.Errorf("failed to decode cache plan update query: %w", err)
		}
		c.Update = &query
	case CachePlanQueryType_DELETE:
		var query CachePlanDeleteQuery
		if err := value.Decode(&query); err != nil {
			return fmt.Errorf("failed to decode cache plan delete query: %w", err)
		}
		c.Delete = &query
	case CachePlanQueryType_INSERT:
		var query CachePlanInsertQuery
		if err := value.Decode(&query); err != nil {
			return fmt.Errorf("failed to decode cache plan insert query: %w", err)
		}
		c.Insert = &query
	default:
		return fmt.Errorf("unknown cache plan query type: %s", base.Type)
	}

	return nil
}

var _ yaml.Unmarshaler = &CachePlanQuery{}

func LoadCachePlan(reader io.Reader) (*CachePlan, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache plan: %w", err)
	}

	var cachePlan CachePlan
	err = yaml.Unmarshal(data, &cachePlan)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal cache plan: %w", err)
	}

	return &cachePlan, nil
}

func (c *CachePlanQuery) MarshalYAML() (interface{}, error) {
	var query interface{}
	switch c.Type {
	case CachePlanQueryType_SELECT:
		query = c.Select
	case CachePlanQueryType_UPDATE:
		query = c.Update
	case CachePlanQueryType_DELETE:
		query = c.Delete
	case CachePlanQueryType_INSERT:
		query = c.Insert
	default:
		return nil, fmt.Errorf("unknown cache plan query type: %s", c.Type)
	}

	return yaml.Marshal(query)
}

var _ yaml.Marshaler = &CachePlanQuery{}

func SaveCachePlan(writer io.Writer, cachePlan *CachePlan) error {
	encoder := yaml.NewEncoder(writer)
	encoder.SetIndent(2)

	if err := encoder.Encode(cachePlan); err != nil {
		return fmt.Errorf("failed to marshal cache plan: %w", err)
	}

	return nil
}
