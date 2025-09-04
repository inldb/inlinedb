package api

import (
	"net/http"

	"github.com/MultiX0/db-test/functions"
	"github.com/MultiX0/db-test/utils"
)

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	result, err := functions.HealthCheck()
	if err != nil {
		utils.RespondError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.WriteJSON(w, http.StatusOK, map[string]any{"ok": result})
}
