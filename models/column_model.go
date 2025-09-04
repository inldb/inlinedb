package models

type ColumnModel struct {
	Name          string  `json:"name"`
	DataType      string  `json:"data_type"`
	IsPrimaryKey  bool    `json:"is_pk"`
	Nullable      bool    `json:"nullable"`
	Default_Value *string `json:"default_value"`
}
