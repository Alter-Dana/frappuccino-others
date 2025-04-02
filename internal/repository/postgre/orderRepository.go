package postgre

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"frappuccino/internal/models"

	"github.com/lib/pq"
)

type orderRepositoryPostgres struct {
	pq     *sql.DB
	logger *slog.Logger
}

func NewOrderRepositoryPostgres(db *sql.DB, logger *slog.Logger) *orderRepositoryPostgres {
	return &orderRepositoryPostgres{
		pq:     db,
		logger: logger,
	}
}

func (m *orderRepositoryPostgres) Insert(order models.Order) (int, error) {
	tx, err := m.pq.Begin()
	if err != nil {
		m.logger.Error("Failed to begin transaction", "error", err)
		return 0, err
	}
	defer tx.Rollback()

	prefsJSON, err := json.Marshal(order.CustomerPreferences)
	if err != nil {
		m.logger.Error("Failed to marshal customer_preferences", "error", err)
		return 0, err
	}

	var orderID int
	err = tx.QueryRow(`INSERT INTO orders (customer_name, order_status, customer_preferences) 
		VALUES ($1, $2, $3) RETURNING id`,
		order.CustomerName, "open", prefsJSON).
		Scan(&orderID)
	if err != nil {
		m.logger.Error(err.Error())
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505":
				return orderID, models.ErrDuplicateOrder
			case "22P02":
				return orderID, models.ErrInvalidEnumTypeInventory
			}
		}
		return orderID, err
	}

	for _, menu := range order.Items {
		_, err = tx.Exec("INSERT INTO order_item (order_id, menu_item_id, quantity) VALUES ($1, $2, $3)",
			orderID, menu.MenuID, menu.Quantity)
		if err != nil {
			m.logger.Error(err.Error())
			if pgErr, ok := err.(*pq.Error); ok {
				switch pgErr.Code {
				case "23503":
					return orderID, models.ErrForeignKeyConstraintOrderMenu
				case "23514":
					return orderID, models.ErrNegativeQuantity
				}
			}
			return 0, err
		}

		rows, err := tx.Query("SELECT inventory_id, quantity FROM menu_item_inventory WHERE menu_id=$1", menu.MenuID)
		if err != nil {
			m.logger.Error(err.Error())
			return orderID, err
		}
		defer rows.Close()

		type invUse struct {
			inventoryID int
			totalNeeded int
		}
		var inventoryList []invUse
		for rows.Next() {
			var inventoryID, perItemQuantity int
			if err := rows.Scan(&inventoryID, &perItemQuantity); err != nil {
				return orderID, err
			}
			inventoryList = append(inventoryList, invUse{inventoryID, perItemQuantity * menu.Quantity})
		}

		for _, item := range inventoryList {
			_, err := tx.Exec("UPDATE inventory SET quantity = quantity - $1 WHERE id = $2", item.totalNeeded, item.inventoryID)
			if err != nil {
				m.logger.Error(err.Error())
				if pqErr, ok := err.(*pq.Error); ok {
					switch pqErr.Code {
					case "23514":
						return orderID, models.ErrNegativeQuantity
					}
				}
				return orderID, err
			}
		}
	}

	return orderID, tx.Commit()
}

func (m *orderRepositoryPostgres) RetrieveAll() ([]models.Order, error) {
	rows, err := m.pq.Query(`
		SELECT o.id, o.customer_name, o.order_status, o.created_at, o.customer_preferences,
		       oi.menu_item_id, oi.quantity
		FROM orders o
		LEFT JOIN order_item oi ON o.id = oi.order_id
		ORDER BY o.id
	`)
	if err != nil {
		m.logger.Error("Failed to execute order query", "error", err)
		return nil, err
	}
	defer rows.Close()

	orderMap := make(map[int]*models.Order)

	for rows.Next() {
		var (
			orderID      int
			customerName string
			status       string
			createdAt    time.Time
			prefsBytes   []byte
			menuItemID   sql.NullInt32
			quantity     sql.NullInt32
		)

		err := rows.Scan(&orderID, &customerName, &status, &createdAt, &prefsBytes, &menuItemID, &quantity)
		if err != nil {
			m.logger.Error("Failed to scan order row", "error", err)
			return nil, err
		}

		if _, ok := orderMap[orderID]; !ok {
			var prefs models.Jsonb
			if err := json.Unmarshal(prefsBytes, &prefs); err != nil {
				m.logger.Error("Failed to unmarshal customer_preferences", "error", err)
				return nil, err
			}

			orderMap[orderID] = &models.Order{
				ID:                  orderID,
				CustomerName:        customerName,
				Status:              status,
				CreatedAt:           createdAt,
				CustomerPreferences: prefs,
				Items:               []models.OrderItem{},
			}
		}

		if menuItemID.Valid {
			orderMap[orderID].Items = append(orderMap[orderID].Items, models.OrderItem{
				MenuID:   int(menuItemID.Int32),
				Quantity: int(quantity.Int32),
			})
		}
	}

	var orders []models.Order
	for _, order := range orderMap {
		orders = append(orders, *order)
	}

	return orders, nil
}

func (m *orderRepositoryPostgres) RetrieveByID(id int) (models.Order, error) {
	rows, err := m.pq.Query(`
		SELECT o.id, o.customer_name, o.order_status, o.created_at, o.customer_preferences,
		       oi.menu_item_id, oi.quantity
		FROM "orders" o
		LEFT JOIN order_item oi ON o.id = oi.order_id
		WHERE o.id = $1
	`, id)
	if err != nil {
		m.logger.Error("Failed to execute order query", "error", err)
		return models.Order{}, err
	}
	defer rows.Close()

	var order models.Order
	for rows.Next() {
		var (
			orderID      int
			customerName string
			status       string
			createdAt    time.Time
			prefsBytes   []byte
			menuItemID   sql.NullInt32
			quantity     sql.NullInt32
		)

		err := rows.Scan(&orderID, &customerName, &status, &createdAt, &prefsBytes, &menuItemID, &quantity)
		if err != nil {
			m.logger.Error("Failed to scan order row", "error", err)
			return models.Order{}, err
		}

		if order.ID == 0 {
			var prefs models.Jsonb
			if err := json.Unmarshal(prefsBytes, &prefs); err != nil {
				m.logger.Error("Failed to unmarshal customer_preferences", "error", err)
				return models.Order{}, err
			}

			order = models.Order{
				ID:                  orderID,
				CustomerName:        customerName,
				Status:              status,
				CreatedAt:           createdAt,
				CustomerPreferences: prefs,
				Items:               []models.OrderItem{},
			}
		}

		if menuItemID.Valid {
			order.Items = append(order.Items, models.OrderItem{
				MenuID:   int(menuItemID.Int32),
				Quantity: int(quantity.Int32),
			})
		}
	}

	if order.ID == 0 {
		return models.Order{}, models.ErrNoRecord
	}

	return order, nil
}

func (m *orderRepositoryPostgres) Update(orderID int, order models.Order) error {
	tx, err := m.pq.Begin()
	if err != nil {
		m.logger.Error("Failed to begin transaction", "error", err)
		return err
	}
	defer tx.Rollback()

	prefsJSON, err := json.Marshal(order.CustomerPreferences)
	if err != nil {
		m.logger.Error("Failed to marshal customer_preferences", "error", err)
		return err
	}

	result, err := tx.Exec(`
		UPDATE orders
		SET customer_name = $1, customer_preferences = $2
		WHERE id = $3 AND order_status=$4
	`, order.CustomerName, prefsJSON, orderID, "open")
	if err != nil {
		m.logger.Error(err.Error())
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505":
				return models.ErrDuplicateOrder
			}
		}
		m.logger.Error("Failed to update order", "error", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		m.logger.Error("Failed to check rows affected", "error", err)
		return err
	}
	if rowsAffected == 0 {
		return models.ErrNoRecord
	}

	_, err = tx.Exec("DELETE FROM order_item WHERE order_id = $1", orderID)
	if err != nil {
		m.logger.Error("Failed to delete order items", "error", err)
		return err
	}

	for _, item := range order.Items {
		_, err = tx.Exec(
			"INSERT INTO order_item (order_id, menu_item_id, quantity) VALUES ($1, $2, $3)",
			orderID, item.MenuID, item.Quantity,
		)
		if err != nil {
			m.logger.Error(err.Error())
			if pgErr, ok := err.(*pq.Error); ok {
				switch pgErr.Code {
				case "23503":
					return models.ErrForeignKeyConstraintOrderMenu
				case "23514":
					return models.ErrNegativeQuantity
				}
			}
			m.logger.Error("Failed to insert order item", "menu_id", item.MenuID, "error", err)
			return err
		}
	}

	return tx.Commit()
}

func (m *orderRepositoryPostgres) Delete(id int) error {
	result, err := m.pq.Exec("DELETE FROM orders WHERE id=$1", id)
	if err != nil {
		m.logger.Error("Failed to execute query", "error", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if rowsAffected == 0 {
		return models.ErrNoRecord
	}

	return err
}

func (m *orderRepositoryPostgres) Close(id int) error {
	result, err := m.pq.Exec(`UPDATE orders SET order_status=$1 WHERE id=$2`, "closed", id)
	if err != nil {
		m.logger.Error("Failed to close order", "error", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		m.logger.Error("Failed to check rows affected", "error", err)
		return err
	}
	if rowsAffected == 0 {
		return models.ErrNoRecord
	}

	return nil
}

func (m *orderRepositoryPostgres) NumberOfOrderedItems(startDate string, endDate string) (map[string]int, error) {
	subquery := "SELECT * FROM orders"
	if startDate != "" || endDate != "" {
		subquery += " WHERE "
		if startDate != "" {
			subquery += fmt.Sprintf("created_at::date >= '%v'", startDate)
			if endDate != "" {
				subquery += fmt.Sprintf(" AND created_at::date <= '%v'", endDate)
			}
		} else {
			subquery += fmt.Sprintf("created_at::date <= '%v'", endDate)
		}
	}
	rows, err := m.pq.Query(fmt.Sprintf(`
		SELECT mi.name as menu_item, SUM(oi.quantity) as quantity
		FROM (%v) o
		JOIN order_item oi ON oi.order_id = o.id
		JOIN menu_items mi ON oi.menu_item_id = mi.id
		GROUP BY mi.id, mi.name
	`, subquery))
	if err != nil {
		m.logger.Error("Failed to execute order query", "error", err)
		return nil, err
	}
	defer rows.Close()

	mp := make(map[string]int)
	for rows.Next() {
		var name string
		var quantity int
		if err = rows.Scan(&name, &quantity); err != nil {
			return nil, err
		}
		mp[name] = quantity
	}
	return mp, nil
}

func (m *orderRepositoryPostgres) GetBatchTotalOrderPrice(orderID int) (float64, error) {
	query := `
		SELECT COALESCE(SUM(oi.quantity * mi.price))
		FROM order_item oi
		JOIN menu_items mi ON oi.menu_item_id = mi.id
		WHERE oi.order_id=$1
	`
	var totalOrderPrice float64
	err := m.pq.QueryRow(query, orderID).Scan(&totalOrderPrice)
	if err != nil {
		m.logger.Error(err.Error())
	}

	return totalOrderPrice, err
}

func (m *orderRepositoryPostgres) GetBatchInventoryUpdates(orderIDs []int) ([]models.BatchInventoryUpdate, error) {
	query := `
		SELECT inv.id, inv.name, SUM(mii.quantity * oi.quantity) AS total_used, inv.quantity AS remaining
		FROM inventory inv
		JOIN menu_item_inventory mii ON inv.id = mii.inventory_id
		JOIN order_item oi ON oi.menu_item_id = mii.menu_id
		JOIN orders o ON o.id = oi.order_id
		WHERE o.id = ANY($1)
		GROUP BY inv.id, inv.name, inv.quantity
	`
	rows, err := m.pq.Query(query, pq.Array(orderIDs))
	if err != nil {
		m.logger.Error("Failed to calculate inventory usage", "error", err)
		return nil, err
	}
	defer rows.Close()

	var updates []models.BatchInventoryUpdate
	for rows.Next() {
		var update models.BatchInventoryUpdate
		err := rows.Scan(&update.ID, &update.Name, &update.QuantityUsed, &update.Remaining)
		if err != nil {
			m.logger.Error("Failed to scan inventory usage row", "error", err)
			return nil, err
		}
		updates = append(updates, update)
	}

	return updates, nil
}
