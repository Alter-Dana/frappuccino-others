package models

import "strconv"

type MenuItem struct {
	ID          int                 `json:"id"`
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Price       float64             `json:"price"`
	Inventory   []MenuItemInventory `json:"inventory"`
}

type MenuItemInventory struct {
	InventoryID int `json:"inventory_id"`
	Quantity    int `json:"quantity"`
}

type menuItemValidator struct {
	errors map[string]string
	menu   MenuItem
}

func NewMenuItemValidator(menu MenuItem) *menuItemValidator {
	return &menuItemValidator{
		errors: make(map[string]string),
		menu:   menu,
	}
}

func (v *menuItemValidator) Validate() map[string]string {
	if v.menu.Name == "" {
		v.errors["Name"] = "Name is required"
	}
	if v.menu.Description == "" {
		v.errors["Description"] = "Description is required"
	}
	if v.menu.Price < 0 {
		v.errors["Price"] = "Price must be 0 or more"
	}
	if len(v.menu.Inventory) < 1 {
		v.errors["Inventory"] = "At least one inventory item is required"
	}

	inventoryIDSet := make(map[int]bool)
	for _, inv := range v.menu.Inventory {
		key := "Inventory[" + strconv.Itoa(inv.InventoryID) + "]"

		if inventoryIDSet[inv.InventoryID] {
			v.errors[key+".InventoryID"] = "Duplicate InventoryID detected"
		} else {
			inventoryIDSet[inv.InventoryID] = true
		}

		if inv.Quantity < 0 {
			v.errors[key+".Quantity"] = "Quantity must be 0 or more"
		}
	}

	if len(v.errors) > 0 {
		return v.errors
	}
	return nil
}
