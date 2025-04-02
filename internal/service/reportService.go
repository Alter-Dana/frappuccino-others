package service

import (
	"database/sql"
	"log/slog"
	"strconv"
	"strings"

	"frappuccino/internal/models"
	"frappuccino/internal/repository"
	"frappuccino/internal/repository/postgre"
	"frappuccino/internal/utils"
)

type reportService struct {
	reportRepo repository.ReportRepository
}

func NewReportService(db *sql.DB, logger *slog.Logger) *reportService {
	return &reportService{postgre.NewReportRepositoryPostgres(db, logger)}
}

func (s *reportService) GetTotalSales() (models.ReportTotalSales, error) {
	report, err := s.reportRepo.GetTotalSales()

	return report, err
}

func (s *reportService) GetPopularMenuItems() ([]models.ReportPopularItem, error) {
	popularItems, err := s.reportRepo.GetPopularMenuItems()

	return popularItems, err
}

func (s *reportService) TextSearch(query string, filter string, minPriceArg string, maxPriceArg string) (models.ReportSearch, error) {
	if filter != "all" && filter != "menu" && filter != "orders" {
		return models.ReportSearch{}, models.ErrInvalidFilterOption
	}
	minPrice, maxPrice, err := utils.ValidatePrices(minPriceArg, maxPriceArg)
	if err != nil {
		return models.ReportSearch{}, err
	}

	var results models.ReportSearch
	results.TotalMatches = 0
	if filter == "menu" || filter == "all" {
		results.MenuResults, err = s.reportRepo.TextSearchMenu(query, minPrice, maxPrice)
		if err != nil {
			return models.ReportSearch{}, err
		}
		results.TotalMatches += len(results.MenuResults)
	}
	if filter == "orders" || filter == "all" {
		results.OrdersResults, err = s.reportRepo.TextSearchOrders(query, minPrice, maxPrice)
		if err != nil {
			return models.ReportSearch{}, err
		}
		results.TotalMatches += len(results.OrdersResults)
	}
	return results, nil
}

func (s *reportService) OrderedItemsByPeriod(period, month, year string) (models.ReportOrderedItems, error) {
	period = strings.ToLower(period)
	month = strings.ToLower(month)
	year = strings.ToLower(year)
	if period != "day" && period != "month" {
		return models.ReportOrderedItems{}, models.ErrInvalidPeriod
	}
	if period == "day" {
		if year != "" {
			return models.ReportOrderedItems{}, models.ErrInvalidOrderedItemsFormat
		}
		monthNum := utils.GetMonthNumber(month)
		if monthNum == -1 {
			return models.ReportOrderedItems{}, models.ErrInvalidOrderedItemsFormat
		}
		data, err := s.reportRepo.OrderedItemsByDays(monthNum)
		if err != nil {
			return models.ReportOrderedItems{}, err
		}
		return models.ReportOrderedItems{
			Period:       period,
			Month:        month,
			OrderedItems: data,
		}, nil
	} // period = month
	if month != "" {
		return models.ReportOrderedItems{}, models.ErrInvalidOrderedItemsFormat
	}
	yearNum, err := strconv.Atoi(year)
	if err != nil || yearNum <= 0 {
		return models.ReportOrderedItems{}, models.ErrInvalidOrderedItemsFormat
	}
	data, err := s.reportRepo.OrderedItemsByMonths(yearNum)
	if err != nil {
		return models.ReportOrderedItems{}, err
	}
	return models.ReportOrderedItems{
		Period:       period,
		Year:         year,
		OrderedItems: data,
	}, nil
}
