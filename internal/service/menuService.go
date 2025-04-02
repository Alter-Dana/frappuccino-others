package service

import (
	"database/sql"
	"log/slog"
	"strconv"

	"frappuccino/internal/models"
	"frappuccino/internal/repository"
	"frappuccino/internal/repository/postgre"
)

type menuService struct {
	menuRepo repository.MenuRepository
}

func NewMenuService(db *sql.DB, logger *slog.Logger) *menuService {
	return &menuService{
		postgre.NewMenuRepositoryPostgres(db, logger),
	}
}

func (s *menuService) InsertMenu(menu models.MenuItem) (map[string]string, error) {
	validator := models.NewMenuItemValidator(menu)
	if errMap := validator.Validate(); errMap != nil {
		return errMap, models.ErrMissingFields
	}

	err := s.menuRepo.InsertMenuItem(menu)
	return nil, err
}

func (s *menuService) RetrieveAll() ([]models.MenuItem, error) {
	menuItems, err := s.menuRepo.RetrieveAll()

	return menuItems, err
}

func (s *menuService) RetrieveByID(id string) (models.MenuItem, error) {
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return models.MenuItem{}, models.ErrInvalidID
	}

	menuItem, err := s.menuRepo.RetrieveByID(idInt)

	return menuItem, err
}

func (s *menuService) Update(id string, menuItem models.MenuItem) (map[string]string, error) {
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return nil, models.ErrInvalidID
	}

	validator := models.NewMenuItemValidator(menuItem)
	if errMap := validator.Validate(); errMap != nil {
		return errMap, models.ErrMissingFields
	}

	err = s.menuRepo.UpdateMenuItem(idInt, menuItem)
	return nil, err
}

func (s *menuService) Delete(id string) error {
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return models.ErrInvalidID
	}

	err = s.menuRepo.Delete(idInt)
	return err
}
