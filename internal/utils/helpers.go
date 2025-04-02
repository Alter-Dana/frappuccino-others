package utils

import (
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"strconv"
	"time"

	"frappuccino/internal/models"
)

// sending responses in the json format
//
//	{
//		"error": "Internal Server Error"
//	}
type Response map[string]interface{}

func SendJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func MapErrorToResponse(err error, validationMap any) (int, any) {
	switch {
	// General errors
	case errors.Is(err, models.ErrInvalidID):
		return http.StatusBadRequest, Response{"error": err.Error()}
	case errors.Is(err, models.ErrNoRecord):
		return http.StatusNotFound, Response{"error": err.Error()}
	case errors.Is(err, models.ErrMissingFields):
		return http.StatusBadRequest, validationMap

	// Inventory errors
	case errors.Is(err, models.ErrDuplicateInventory),
		errors.Is(err, models.ErrNegativeQuantity),
		errors.Is(err, models.ErrInvalidEnumTypeInventory):
		return http.StatusBadRequest, Response{"error": err.Error()}

	// Menu errors
	case errors.Is(err, models.ErrDuplicateMenuItem),
		errors.Is(err, models.ErrNegativePrice),
		errors.Is(err, models.ErrForeignKeyConstraintMenuInventory):
		return http.StatusBadRequest, Response{"error": err.Error()}

	// Order errors
	case errors.Is(err, models.ErrDuplicateOrder),
		errors.Is(err, models.ErrInvalidFilterOption),
		errors.Is(err, models.ErrForeignKeyConstraintOrderMenu):
		return http.StatusBadRequest, Response{"error": err.Error()}

	// Report errors
	case errors.Is(err, models.ErrInvalidPrice),
		errors.Is(err, models.ErrInvalidPeriod),
		errors.Is(err, models.ErrInvalidOrderedItemsFormat):
		return http.StatusBadRequest, Response{"error": err.Error()}

	// Default catch-all
	default:
		return http.StatusInternalServerError, Response{"error": "Internal Server Error"}
	}
}

func ValidatePrices(minPriceStr string, maxPriceStr string) (float64, float64, error) {
	var minPrice, maxPrice float64
	var err error

	if minPriceStr == "absent" {
		minPrice = -1
	} else {
		minPrice, err = strconv.ParseFloat(minPriceStr, 64)
		if err != nil || math.IsNaN(minPrice) || math.IsInf(minPrice, 0) {
			return 0, 0, models.ErrInvalidPrice
		}
	}
	if maxPriceStr == "absent" {
		maxPrice = -1
	} else {
		maxPrice, err = strconv.ParseFloat(maxPriceStr, 64)
		if err != nil || math.IsNaN(maxPrice) || math.IsInf(maxPrice, 0) {
			return 0, 0, models.ErrInvalidPrice
		}
	}
	if minPrice != -1 && maxPrice != -1 && minPrice > maxPrice {
		return 0, 0, models.ErrInvalidPrice
	}
	return minPrice, maxPrice, nil
}

func ConvertDateFormat(dateStr string) string {
	date, err := time.Parse("02.01.2006", dateStr)
	if err == nil {
		return date.Format("2006-01-02")
	}
	date, err = time.Parse("2006-01-02", dateStr)
	if err == nil {
		return date.Format("2006-01-02")
	}
	return ""
}

func GetMonthNumber(month string) int {
	monthNum := map[string]int{
		"january":   1,
		"february":  2,
		"march":     3,
		"april":     4,
		"may":       5,
		"june":      6,
		"july":      7,
		"august":    8,
		"september": 9,
		"october":   10,
		"november":  11,
		"december":  12,
	}
	num, ok := monthNum[month]
	if !ok {
		return -1
	}
	return num
}

func GetMonthName(month int) string {
	months := []string{
		"january", "february", "march", "april",
		"may", "june", "july", "august",
		"september", "october", "november", "december",
	}
	if month < 1 || month > 12 {
		return ""
	}
	return months[month-1]
}

func GetDaysInMonth(month int) int {
	days := []int{31, 29, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	if month < 1 || month > 12 {
		return -1
	}
	return days[month-1]
}
