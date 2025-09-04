package api

import (
	"encoding/json"
	"net/http"

	"github.com/MultiX0/db-test/functions"
	"github.com/MultiX0/db-test/utils"
)

func RowsAsJson(w http.ResponseWriter, r *http.Request) {
	type BodyStruct struct {
		QUERY string `json:"query"`
	}

	var body BodyStruct

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	data, err := functions.RawSQL(body.QUERY)
	if err != nil {
		utils.RespondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if bytes, ok := data.([]byte); ok {
		var dataJson []map[string]any
		err = json.Unmarshal(bytes, &dataJson)
		if err != nil {
			utils.RespondError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		utils.WriteJSON(w, http.StatusOK, map[string]any{"data": dataJson})
	} else {
		utils.WriteJSON(w, http.StatusOK, data)
		return
	}

}
