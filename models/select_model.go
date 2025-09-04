package models

type SelectModel struct {
	TableName       string        `json:"table"`
	SelectedColumns []string      `json:"columns"`
	Filters         []FilterGroup `json:"filters"`
}

type FilterGroup struct {
	Conditions []FilterCondition `json:"conditions"`
	Logic      string            `json:"logic"` // AND, OR
}
type FilterCondition struct {
	Column   string `json:"column"`
	Operator string `json:"operator"` // eq, ne, gt, lt, gte, lte, like, in, not_in
	Value    any    `json:"value"`
}
