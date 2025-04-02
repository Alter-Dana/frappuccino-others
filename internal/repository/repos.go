package repository

import "frappuccino/internal/models"

type InventoryRepository interface {
	Insert(name, unit string, quantity int, categories []string) error
	RetrieveByID(id int) (models.Inventory, error)
	RetrieveAll() ([]models.Inventory, error)
	Update(id int, name, unit string, quantity int, categories []string) error
	Delete(id int) error
	GetLeftOvers(sortBy string, page, pageSize int) ([]models.InventoryLeftOverItem, int, error)
}

type MenuRepository interface {
	InsertMenuItem(item models.MenuItem) error
	RetrieveAll() ([]models.MenuItem, error)
	RetrieveByID(id int) (models.MenuItem, error)
	UpdateMenuItem(menuID int, menuItem models.MenuItem) error
	Delete(id int) error
}

type OrderRepository interface {
	Insert(order models.Order) (int, error)
	RetrieveAll() ([]models.Order, error)
	RetrieveByID(id int) (models.Order, error)
	Update(orderID int, order models.Order) error
	Delete(id int) error
	Close(id int) error
	NumberOfOrderedItems(startDate string, endDate string) (map[string]int, error)
	GetBatchTotalOrderPrice(orderID int) (float64, error)
	GetBatchInventoryUpdates(orderIDs []int) ([]models.BatchInventoryUpdate, error)
}

type ReportRepository interface {
	GetTotalSales() (models.ReportTotalSales, error)
	GetPopularMenuItems() ([]models.ReportPopularItem, error)
	TextSearchMenu(query string, minPrice float64, maxPrice float64) ([]models.ReportMenuSearchItem, error)
	TextSearchOrders(query string, minPrice float64, maxPrice float64) ([]models.ReportOrderSearchItem, error)
	OrderedItemsByDays(month int) ([]map[string]int, error)
	OrderedItemsByMonths(year int) ([]map[string]int, error)
}
