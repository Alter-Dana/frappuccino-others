package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"frappuccino/internal/models"
	"frappuccino/internal/utils"
)

func (app *application) menuCreate(w http.ResponseWriter, r *http.Request) {
	var menuItem models.MenuItem
	err := json.NewDecoder(r.Body).Decode(&menuItem)
	if err != nil {
		utils.SendJSONResponse(w, http.StatusBadRequest, utils.Response{"error": "request body does not match json format"})
		return
	}
	defer r.Body.Close()

	m, err := app.MenuSvc.InsertMenu(menuItem)
	if err != nil {
		status, body := utils.MapErrorToResponse(err, m)
		utils.SendJSONResponse(w, status, body)
		return
	}

	utils.SendJSONResponse(w, http.StatusCreated, utils.Response{"message": "created"})
}

func (app *application) menuRetrieveAll(w http.ResponseWriter, r *http.Request) {
	menuItems, err := app.MenuSvc.RetrieveAll()
	if err != nil {
		utils.SendJSONResponse(w, http.StatusInternalServerError, utils.Response{"error": "Internal Server Error"})
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, menuItems)
}

func (app *application) menuRetrieveAllByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	menuItem, err := app.MenuSvc.RetrieveByID(id)
	if err != nil {
		status, body := utils.MapErrorToResponse(err, nil)
		utils.SendJSONResponse(w, status, body)
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, menuItem)
}

func (app *application) menuUpdate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var menuItem models.MenuItem
	err := json.NewDecoder(r.Body).Decode(&menuItem)
	if err != nil {
		utils.SendJSONResponse(w, http.StatusBadRequest, utils.Response{"error": "request body does not match json format"})
		return
	}
	defer r.Body.Close()

	m, err := app.MenuSvc.Update(id, menuItem)
	if err != nil {
		status, body := utils.MapErrorToResponse(err, m)
		utils.SendJSONResponse(w, status, body)
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, utils.Response{"message": fmt.Sprintf("Updated menu item %s", id)})
}

func (app *application) menuDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	err := app.MenuSvc.Delete(id)
	if err != nil {
		status, body := utils.MapErrorToResponse(err, nil)
		utils.SendJSONResponse(w, status, body)
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, utils.Response{"message": fmt.Sprintf("Deleted %s", id)})
}
