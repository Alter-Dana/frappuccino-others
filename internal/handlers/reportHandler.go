package handlers

import (
	"net/http"

	"frappuccino/internal/utils"
)

func (app *application) getTotalSalesReport(w http.ResponseWriter, r *http.Request) {
	report, err := app.ReportSvc.GetTotalSales()
	if err != nil {
		status, body := utils.MapErrorToResponse(err, nil)
		utils.SendJSONResponse(w, status, body)
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, report)
}

func (app *application) getPopularMenuItems(w http.ResponseWriter, r *http.Request) {
	popularItems, err := app.ReportSvc.GetPopularMenuItems()
	if err != nil {
		status, body := utils.MapErrorToResponse(err, nil)
		utils.SendJSONResponse(w, status, body)
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, popularItems)
}

func (app *application) textSearch(w http.ResponseWriter, r *http.Request) {
	queryArgs := r.URL.Query()
	textQueryArr := queryArgs["q"]
	if len(textQueryArr) == 0 || len(textQueryArr[0]) == 0 {
		utils.SendJSONResponse(w, http.StatusBadRequest, utils.Response{"error": "text search query is invalid"})
		return
	}
	args := []string{textQueryArr[0]}
	for i, arg := range [][]string{queryArgs["filter"], queryArgs["minPrice"], queryArgs["maxPrice"]} {
		if len(arg) > 0 {
			args = append(args, arg[0])
		} else {
			if i == 0 {
				args = append(args, "all")
			} else {
				args = append(args, "absent")
			}
		}
	}
	data, err := app.ReportSvc.TextSearch(args[0], args[1], args[2], args[3])
	if err != nil {
		status, body := utils.MapErrorToResponse(err, nil)
		utils.SendJSONResponse(w, status, body)
		return
	}
	utils.SendJSONResponse(w, http.StatusOK, data)
}

func (app *application) orderedItemsByPeriod(w http.ResponseWriter, r *http.Request) {
	queryArgs := r.URL.Query()
	periodArr := queryArgs["period"]
	if len(periodArr) == 0 {
		utils.SendJSONResponse(w, http.StatusBadRequest, utils.Response{"error": "period must be specified"})
		return
	}
	monthArr := queryArgs["month"]
	yearArr := queryArgs["year"]
	if len(monthArr) == 0 {
		monthArr = append(monthArr, "")
	}
	if len(yearArr) == 0 {
		yearArr = append(yearArr, "")
	}
	data, err := app.ReportSvc.OrderedItemsByPeriod(periodArr[0], monthArr[0], yearArr[0])
	if err != nil {
		status, body := utils.MapErrorToResponse(err, nil)
		utils.SendJSONResponse(w, status, body)
		return
	}
	utils.SendJSONResponse(w, http.StatusOK, data)
}
