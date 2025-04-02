package models

type Inventory struct {
	ID         int      `json:"id"`
	Name       string   `json:"name"`
	Quantity   int      `json:"quantity"`
	Unit       string   `json:"unit"`
	Categories []string `json:"categories"`
}

type inventoryValidator struct {
	validator map[string]string
	inventory Inventory
}

func NewInventoryValidator(inventory Inventory) *inventoryValidator {
	return &inventoryValidator{
		make(map[string]string),
		inventory,
	}
}

func (v *inventoryValidator) Validate() map[string]string {
	if v.inventory.Name == "" {
		v.validator["Name"] = "missing Name"
	}
	if v.inventory.Unit == "" {
		v.validator["Unit"] = "missing Unit"
	}
	if v.inventory.Quantity < 0 {
		v.validator["Quantity"] = "Quantity must be 0 or more"
	}

	if len(v.validator) > 0 {
		return v.validator
	}
	return nil
}

type InventoryLeftOverItem struct {
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
}

type InventoryLeftOversResponse struct {
	CurrentPage int                     `json:"currentPage"`
	HasNextPage bool                    `json:"hasNextPage"`
	PageSize    int                     `json:"pageSize"`
	TotalPages  int                     `json:"totalPages"`
	Data        []InventoryLeftOverItem `json:"data"`
}
