package postgre

import (
	"database/sql"
	"fmt"
	"log/slog"
	"strconv"

	"frappuccino/internal/models"
	"frappuccino/internal/utils"

	"github.com/lib/pq"
)

type reportRepositoryPostgres struct {
	pq     *sql.DB
	logger *slog.Logger
}

func NewReportRepositoryPostgres(db *sql.DB, logger *slog.Logger) *reportRepositoryPostgres {
	return &reportRepositoryPostgres{
		pq:     db,
		logger: logger,
	}
}

func (m *reportRepositoryPostgres) GetTotalSales() (models.ReportTotalSales, error) {
	query := `
		SELECT
			COUNT(DISTINCT o.id) AS orders_completed,
			COALESCE(SUM(oi.quantity * mi.price), 0) AS total_sales
		FROM orders o
		JOIN order_item oi ON o.id = oi.order_id
		JOIN menu_items mi ON oi.menu_item_id = mi.id
		WHERE o.order_status = 'closed';
	`
	var report models.ReportTotalSales
	err := m.pq.QueryRow(query).Scan(&report.OrdersCompleted, &report.TotalSales)
	if err != nil {
		m.logger.Error(err.Error())
		return models.ReportTotalSales{}, err
	}

	return report, nil
}

func (m *reportRepositoryPostgres) GetPopularMenuItems() ([]models.ReportPopularItem, error) {
	query := `
		SELECT
			mi.name,
			mi.description,
			mi.price,
			SUM(oi.quantity) AS total_items_sold,
			RANK() OVER (ORDER BY SUM(oi.quantity) DESC) AS rank
		FROM order_item oi
		JOIN menu_items mi ON oi.menu_item_id = mi.id
		JOIN orders o ON oi.order_id = o.id
		WHERE o.order_status = 'closed'
		GROUP BY mi.id
		ORDER BY total_items_sold DESC
		LIMIT 5;
	`
	rows, err := m.pq.Query(query)
	if err != nil {
		m.logger.Error(err.Error())
		return nil, err
	}
	defer rows.Close()

	var popularItems []models.ReportPopularItem
	for rows.Next() {
		var item models.ReportPopularItem

		err = rows.Scan(&item.Name, &item.Description, &item.Price, &item.TotalItemsSold, &item.Rank)
		if err != nil {
			m.logger.Error(err.Error())
			return nil, err
		}

		popularItems = append(popularItems, item)
	}

	return popularItems, nil
}

func (m *reportRepositoryPostgres) TextSearchMenu(query string, minPrice float64, maxPrice float64) ([]models.ReportMenuSearchItem, error) {
	queryArgs := []any{query}
	dbQuery := `
		WITH q AS (
			SELECT plainto_tsquery('english', $1) as q
		)
		SELECT id, name, description, price, ts_rank(tsv, q.q) as relevance
		FROM menu_items
		CROSS JOIN q
		WHERE tsv @@ q.q`
	if minPrice != -1 {
		fmt.Println(minPrice)
		queryArgs = append(queryArgs, minPrice)
		dbQuery += fmt.Sprintf(" AND price >= $%v", len(queryArgs))
	}
	if maxPrice != -1 {
		fmt.Println(maxPrice)
		queryArgs = append(queryArgs, maxPrice)
		dbQuery += fmt.Sprintf(" AND price <= $%v", len(queryArgs))
	}
	dbQuery += "\nORDER BY relevance desc;"

	rows, err := m.pq.Query(dbQuery, queryArgs...)
	if err != nil {
		m.logger.Error(err.Error())
		return nil, err
	}
	defer rows.Close()

	var results []models.ReportMenuSearchItem
	for rows.Next() {
		var resItem models.ReportMenuSearchItem
		err = rows.Scan(&resItem.Id, &resItem.Name, &resItem.Description, &resItem.Price, &resItem.Relevance)
		if err != nil {
			m.logger.Error(err.Error())
			return nil, err
		}
		results = append(results, resItem)
	}
	return results, nil
}

func (m *reportRepositoryPostgres) TextSearchOrders(query string, minPrice float64, maxPrice float64) ([]models.ReportOrderSearchItem, error) {
	queryArgs := []any{query}
	dbQuery := `
		WITH q AS (
			SELECT plainto_tsquery('english', $1) as q
		)
		SELECT o.id, o.customer_name, array_agg(mi.name), SUM(oi.quantity * mi.price), MAX(ts_rank(mi.tsv, q.q)) as relevance
		FROM menu_items mi
		CROSS JOIN q
		JOIN order_item oi ON mi.id = oi.menu_item_id
		JOIN orders o ON o.id = oi.order_id
		WHERE mi.tsv @@ q.q`
	if minPrice != -1 {
		queryArgs = append(queryArgs, minPrice)
		dbQuery += fmt.Sprintf(" AND mi.price >= $%v", len(queryArgs))
	}
	if maxPrice != -1 {
		queryArgs = append(queryArgs, maxPrice)
		dbQuery += fmt.Sprintf(" AND mi.price <= $%v", len(queryArgs))
	}
	dbQuery += `
		GROUP BY o.id, o.customer_name
		ORDER BY relevance desc;`

	rows, err := m.pq.Query(dbQuery, queryArgs...)
	if err != nil {
		m.logger.Error(err.Error())
		return nil, err
	}
	defer rows.Close()

	var results []models.ReportOrderSearchItem
	for rows.Next() {
		var resItem models.ReportOrderSearchItem
		err = rows.Scan(&resItem.Id, &resItem.CustomerName, pq.Array(&resItem.Items), &resItem.Total, &resItem.Relevance)
		if err != nil {
			m.logger.Error(err.Error())
			return nil, err
		}
		results = append(results, resItem)
	}
	return results, nil
}

func (m *reportRepositoryPostgres) OrderedItemsByDays(month int) ([]map[string]int, error) {
	rows, err := m.pq.Query(`
		SELECT EXTRACT(DAY FROM created_at) AS day, COUNT(id) AS num
		FROM orders
		WHERE EXTRACT(MONTH FROM created_at)=$1
		GROUP BY day`, month)
	if err != nil {
		m.logger.Error(err.Error())
		return nil, err
	}
	defer rows.Close()

	results := make([]map[string]int, utils.GetDaysInMonth(month))
	for i := range results {
		results[i] = map[string]int{strconv.Itoa(i + 1): 0}
	}
	for rows.Next() {
		var day, cnt int
		err = rows.Scan(&day, &cnt)
		if err != nil {
			m.logger.Error(err.Error())
			return nil, err
		}
		results[day-1][strconv.Itoa(day)] = cnt
	}
	return results, nil
}

func (m *reportRepositoryPostgres) OrderedItemsByMonths(year int) ([]map[string]int, error) {
	rows, err := m.pq.Query(`
		SELECT EXTRACT(MONTH FROM created_at) AS month, COUNT(id) AS num
		FROM orders
		WHERE EXTRACT(YEAR FROM created_at)=$1
		GROUP BY month`, year)
	if err != nil {
		m.logger.Error(err.Error())
		return nil, err
	}
	defer rows.Close()

	results := make([]map[string]int, 12)
	for i := range results {
		results[i] = map[string]int{utils.GetMonthName(i + 1): 0}
	}
	for rows.Next() {
		var mon, cnt int
		err = rows.Scan(&mon, &cnt)
		if err != nil {
			m.logger.Error(err.Error())
			return nil, err
		}
		results[mon-1][utils.GetMonthName(mon)] = cnt
	}
	return results, nil
}
