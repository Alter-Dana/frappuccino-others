package postgre

import (
	"database/sql"
	"errors"
	"log/slog"

	"frappuccino/internal/models"

	"github.com/lib/pq"
)

type menuRepositoryPostgres struct {
	pq     *sql.DB
	logger *slog.Logger
}

func NewMenuRepositoryPostgres(db *sql.DB, logger *slog.Logger) *menuRepositoryPostgres {
	return &menuRepositoryPostgres{
		pq:     db,
		logger: logger,
	}
}

func (m *menuRepositoryPostgres) InsertMenuItem(menuItem models.MenuItem) error {
	tx, err := m.pq.Begin()
	if err != nil {
		m.logger.Error("Failed to begin transaction", "error", err)
		return err
	}
	defer tx.Rollback()

	var menuID int
	err = tx.QueryRow(`INSERT INTO menu_items (name, description, price) VALUES ($1, $2, $3) RETURNING id`,
		menuItem.Name, menuItem.Description, menuItem.Price).
		Scan(&menuID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505":
				return models.ErrDuplicateMenuItem
			case "23514":
				return models.ErrNegativePrice
			}
		}
		return err
	}

	for _, inv := range menuItem.Inventory {
		_, err = tx.Exec(
			"INSERT INTO menu_item_inventory (menu_id, inventory_id, quantity) VALUES ($1, $2, $3)",
			menuID, inv.InventoryID, inv.Quantity,
		)
		if err != nil {
			if pgErr, ok := err.(*pq.Error); ok {
				switch pgErr.Code {
				case "23503":
					return models.ErrForeignKeyConstraintMenuInventory
				case "23514":
					return models.ErrNegativeQuantity
				}
			}
			return err
		}
	}

	return tx.Commit()
}

func (m *menuRepositoryPostgres) RetrieveAll() ([]models.MenuItem, error) {
	rows, err := m.pq.Query(`
		SELECT menu.id, menu.name, menu.description, menu.price, inventory.inventory_id, inventory.quantity
		FROM menu_items AS menu
		LEFT JOIN menu_item_inventory AS inventory
		ON menu.id=inventory.menu_id
	`)
	if err != nil {
		m.logger.Error("Failed to execute Query", "error", err)
		return nil, err
	}
	defer rows.Close()

	menuMap := make(map[int]*models.MenuItem)
	for rows.Next() {
		var id int
		var name, description string
		var price float64
		var inventoryID, quantity sql.NullInt32

		err := rows.Scan(&id, &name, &description, &price, &inventoryID, &quantity)
		if err != nil {
			m.logger.Error("Failed to scan row", "error", err)
			return nil, err
		}

		if _, ok := menuMap[id]; !ok {
			menuMap[id] = &models.MenuItem{
				ID:          id,
				Name:        name,
				Description: description,
				Price:       price,
				Inventory:   []models.MenuItemInventory{},
			}
		}

		if inventoryID.Valid {
			menuMap[id].Inventory = append(menuMap[id].Inventory, models.MenuItemInventory{
				InventoryID: int(inventoryID.Int32),
				Quantity:    int(quantity.Int32),
			})
		}
	}

	var menuItems []models.MenuItem
	for _, menu := range menuMap {
		menuItems = append(menuItems, *menu)
	}

	return menuItems, nil
}

func (m *menuRepositoryPostgres) RetrieveByID(id int) (models.MenuItem, error) {
	rows, err := m.pq.Query(`
		SELECT menu.id, menu.name, menu.description, menu.price, 
		       inventory.inventory_id, inventory.quantity
		FROM menu_items AS menu
		LEFT JOIN menu_item_inventory AS inventory
		ON menu.id = inventory.menu_id
		WHERE menu.id = $1
	`, id)
	if err != nil {
		m.logger.Error("Failed to execute query", "error", err)
		return models.MenuItem{}, err
	}
	defer rows.Close()

	var menuItem models.MenuItem
	for rows.Next() {
		var inventoryID, quantity sql.NullInt32

		err = rows.Scan(
			&menuItem.ID,
			&menuItem.Name,
			&menuItem.Description,
			&menuItem.Price,
			&inventoryID,
			&quantity,
		)
		if err != nil {
			m.logger.Error("Failed to scan row", "error", err)
			return models.MenuItem{}, err
		}

		if inventoryID.Valid {
			menuItem.Inventory = append(menuItem.Inventory, models.MenuItemInventory{
				InventoryID: int(inventoryID.Int32),
				Quantity:    int(quantity.Int32),
			})
		}
	}

	if menuItem.ID == 0 {
		return models.MenuItem{}, models.ErrNoRecord
	}

	return menuItem, nil
}

func (m *menuRepositoryPostgres) UpdateMenuItem(menuID int, menuItem models.MenuItem) error {
	tx, err := m.pq.Begin()
	if err != nil {
		m.logger.Error("Failed to begin transaction", "error", err)
		return err
	}
	defer tx.Rollback()

	result, err := tx.Exec(`
		UPDATE menu_items
		SET name = $1, description = $2, price = $3 WHERE id = $4
	`, menuItem.Name, menuItem.Description, menuItem.Price, menuID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.ErrNoRecord
		}
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505":
				return models.ErrDuplicateMenuItem
			case "23514":
				return models.ErrNegativePrice
			}
		}
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if rowsAffected == 0 {
		return models.ErrNoRecord
	}

	_, err = tx.Exec("DELETE FROM menu_item_inventory WHERE menu_id = $1", menuID)
	if err != nil {
		m.logger.Error("Failed to execute query", "error", err)
		return err
	}

	for _, inv := range menuItem.Inventory {
		_, err = tx.Exec(
			"INSERT INTO menu_item_inventory (menu_id, inventory_id, quantity) VALUES ($1, $2, $3)",
			menuID, inv.InventoryID, inv.Quantity,
		)
		if err != nil {
			if pgErr, ok := err.(*pq.Error); ok {
				switch pgErr.Code {
				case "23503":
					return models.ErrForeignKeyConstraintMenuInventory
				case "23514":
					return models.ErrNegativeQuantity
				}
			}
			return err
		}
	}

	return tx.Commit()
}

func (m *menuRepositoryPostgres) Delete(id int) error {
	stmt := "DELETE FROM menu_items WHERE id = $1"
	result, err := m.pq.Exec(stmt, id)
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
