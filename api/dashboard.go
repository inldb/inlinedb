package api

import (
	"net/http"

	"github.com/MultiX0/db-test/functions"
	"github.com/MultiX0/db-test/utils"
)

func GetOverview(w http.ResponseWriter, r *http.Request) {

	count, err := functions.GetNumberOfTables()
	if err != nil {
		utils.RespondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if count == nil {
		utils.WriteJSON(w, http.StatusOK, map[string]any{
			"count": 0,
		})
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{
		"count": count,
	})

}
