package service

import (
	"database/sql"
	"log/slog"
	"strconv"

	"frappuccino/internal/models"
	"frappuccino/internal/repository"
	"frappuccino/internal/repository/postgre"
)

type inventoryService struct {
	inventoryRepo repository.InventoryRepository
}

func NewInventoryService(db *sql.DB, logger *slog.Logger) *inventoryService {
	return &inventoryService{
		inventoryRepo: postgre.NewInventoryRepositoryWithPostgres(db, logger),
	}
}

func (s *inventoryService) Insert(inventory models.Inventory) (map[string]string, error) {
	validator := models.NewInventoryValidator(inventory)
	m := validator.Validate()
	if m != nil {
		return m, models.ErrMissingFields
	}

	err := s.inventoryRepo.Insert(inventory.Name, inventory.Unit, inventory.Quantity, inventory.Categories)

	return nil, err
}

func (s *inventoryService) RetrieveByID(id string) (models.Inventory, error) {
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return models.Inventory{}, models.ErrInvalidID
	}

	inventory, err := s.inventoryRepo.RetrieveByID(idInt)

	return inventory, err
}

func (s *inventoryService) RetrieveAll() ([]models.Inventory, error) {
	inventory, err := s.inventoryRepo.RetrieveAll()

	return inventory, err
}

func (s *inventoryService) Update(inventory models.Inventory, id string) (map[string]string, error) {
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return nil, models.ErrInvalidID
	}

	validator := models.NewInventoryValidator(inventory)
	m := validator.Validate()
	if len(m) > 0 {
		return m, models.ErrMissingFields
	}

	err = s.inventoryRepo.Update(idInt, inventory.Name, inventory.Unit, inventory.Quantity, inventory.Categories)
	return nil, err
}

func (s *inventoryService) Delete(id string) error {
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return models.ErrInvalidID
	}

	err = s.inventoryRepo.Delete(idInt)
	return err
}

func (s *inventoryService) GetLeftOvers(sortBy string, page, pageSize int) (models.InventoryLeftOversResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	var sortColumn string
	switch sortBy {
	case "name":
		sortColumn = "name"
	case "quantity", "":
		sortColumn = "quantity"
	default:
		sortColumn = "quantity"
	}

	data, totalPages, err := s.inventoryRepo.GetLeftOvers(sortColumn, page, pageSize)
	if err != nil {
		return models.InventoryLeftOversResponse{}, err
	}

	hasNext := page < totalPages

	return models.InventoryLeftOversResponse{
		CurrentPage: page,
		PageSize:    pageSize,
		TotalPages:  totalPages,
		HasNextPage: hasNext,
		Data:        data,
	}, nil
}
