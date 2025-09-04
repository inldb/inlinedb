package models

type TablesModel struct {
	Tables []TableModel `json:"tables"`
}

type TableModel struct {
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	Sql          string        `json:"sql"`
	Columns      []ColumnModel `json:"columns"`
	RecordsCount int           `json:"records_count"`
}
