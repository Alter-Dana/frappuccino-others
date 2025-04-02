package models

import "errors"

var (
	ErrNoRecord         = errors.New("models: no record")
	ErrNegativeQuantity = errors.New("models: positive quantity constraint violation")
	ErrNegativePrice    = errors.New("models: positive price constraint violation")
	ErrMissingFields    = errors.New("models: missing fields")
	ErrInvalidID        = errors.New("id is not valid int")

	// Inventory errors
	ErrDuplicateInventory       = errors.New("models: duplicate inventory")
	ErrInvalidEnumTypeInventory = errors.New("models: invalid enum type. Supported types: shots, ml, g, units")

	// Menu errors
	ErrDuplicateMenuItem                 = errors.New("models: duplicate menu item")
	ErrForeignKeyConstraintMenuInventory = errors.New("inventory does not exist")

	// Order errors
	ErrDuplicateOrder                = errors.New("models: duplicate order")
	ErrForeignKeyConstraintOrderMenu = errors.New("menu item does not exist")
	ErrInvalidFilterOption           = errors.New("wrong filter option chosen (should be menu/order/all)")

	// Report errors
	ErrInvalidPrice              = errors.New("invalid min/max prices given")
	ErrInvalidPeriod             = errors.New("invalid period type; should be 'day' or 'month'")
	ErrInvalidOrderedItemsFormat = errors.New("invalid format for ordered items by period: should be day/month or month/year")
)
