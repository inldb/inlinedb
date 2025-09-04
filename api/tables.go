package api

import (
	"encoding/json"
	"net/http"

	"github.com/MultiX0/db-test/functions"
	"github.com/MultiX0/db-test/models"
	"github.com/MultiX0/db-test/utils"
)

func GetAllTables(w http.ResponseWriter, r *http.Request) {
	tables, err := functions.GetAllTables()
	if err != nil {
		utils.RespondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusOK, tables.Tables)
}

func GetTable(w http.ResponseWriter, r *http.Request) {
	tableName := r.URL.Query().Get("name")
	if len(tableName) == 0 {
		GetAllTables(w, r)
		return
	}

	table, err := functions.GetTableData(tableName)
	if err != nil {
		utils.RespondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusOK, table)

}

func CreateTable(w http.ResponseWriter, r *http.Request) {

	var table models.TableModel
	if err := json.NewDecoder(r.Body).Decode(&table); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	err := functions.CreateTable(table)
	if err != nil {
		utils.RespondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_table, err := functions.GetTableData(table.Name)
	if err != nil {
		utils.RespondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusOK, _table)

}

func InsertIntoTable(w http.ResponseWriter, r *http.Request) {
	var insertModel models.InsertModel
	if err := json.NewDecoder(r.Body).Decode(&insertModel); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := functions.InsertIntoTable(insertModel)
	if err != nil {
		utils.RespondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"msg": "success",
		"id":  id,
	})

}
