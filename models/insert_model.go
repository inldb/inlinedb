package models

type InsertModel struct {
	TableName string   `json:"table"`
	Columns   []string `json:"columns"`
	Values    []any    `json:"values"`
}
