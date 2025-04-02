package service

import "frappuccino/internal/models"

type InventoryService interface {
	Insert(inventory models.Inventory) (map[string]string, error)
	RetrieveByID(id string) (models.Inventory, error)
	RetrieveAll() ([]models.Inventory, error)
	Update(inventory models.Inventory, id string) (map[string]string, error)
	Delete(id string) error
	GetLeftOvers(sortBy string, page, pageSize int) (models.InventoryLeftOversResponse, error)
}

type MenuService interface {
	InsertMenu(menuItem models.MenuItem) (map[string]string, error)
	RetrieveAll() ([]models.MenuItem, error)
	RetrieveByID(id string) (models.MenuItem, error)
	Update(id string, menuItem models.MenuItem) (map[string]string, error)
	Delete(id string) error
}

type OrderService interface {
	Insert(order models.Order) (map[string]string, error)
	RetrieveAll() ([]models.Order, error)
	RetrieveByID(id string) (models.Order, error)
	Update(id string, order models.Order) (map[string]string, error)
	Delete(id string) error
	Close(id string) error
	NumberOfOrderedItems(startDate string, endDate string) (map[string]int, error)
	BatchOrderProcess(orders []models.Order) (models.BatchOrderResponse, error)
}

type ReportService interface {
	GetTotalSales() (models.ReportTotalSales, error)
	GetPopularMenuItems() ([]models.ReportPopularItem, error)
	TextSearch(query string, filter string, minPriceArg string, maxPriceArg string) (models.ReportSearch, error)
	OrderedItemsByPeriod(period, month, year string) (models.ReportOrderedItems, error)
}
