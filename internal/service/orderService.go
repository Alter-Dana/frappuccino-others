package service

import (
	"database/sql"
	"errors"
	"log/slog"
	"strconv"

	"frappuccino/internal/models"
	"frappuccino/internal/repository"
	"frappuccino/internal/repository/postgre"
	"frappuccino/internal/utils"
)

type orderService struct {
	orderRepo repository.OrderRepository
}

func NewOrderService(db *sql.DB, logger *slog.Logger) *orderService {
	return &orderService{
		postgre.NewOrderRepositoryPostgres(db, logger),
	}
}

func (s *orderService) Insert(order models.Order) (map[string]string, error) {
	validator := models.NewOrderValidator(order)
	if errMap := validator.Validate(); errMap != nil {
		return errMap, models.ErrMissingFields
	}

	_, err := s.orderRepo.Insert(order)
	return nil, err
}

func (s *orderService) RetrieveAll() ([]models.Order, error) {
	orders, err := s.orderRepo.RetrieveAll()

	return orders, err
}

func (s *orderService) RetrieveByID(id string) (models.Order, error) {
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return models.Order{}, models.ErrInvalidID
	}

	order, err := s.orderRepo.RetrieveByID(idInt)

	return order, err
}

func (s *orderService) Update(id string, order models.Order) (map[string]string, error) {
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return nil, models.ErrInvalidID
	}

	validator := models.NewOrderValidator(order)
	if errMap := validator.Validate(); errMap != nil {
		return errMap, models.ErrMissingFields
	}

	return nil, s.orderRepo.Update(idInt, order)
}

func (s *orderService) Delete(id string) error {
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return models.ErrInvalidID
	}

	return s.orderRepo.Delete(idInt)
}

func (s *orderService) Close(id string) error {
	idInt, err := strconv.Atoi(id)
	if err != nil {
		return models.ErrInvalidID
	}

	return s.orderRepo.Close(idInt)
}

func (s *orderService) NumberOfOrderedItems(startDate string, endDate string) (map[string]int, error) {
	return s.orderRepo.NumberOfOrderedItems(
		utils.ConvertDateFormat(startDate),
		utils.ConvertDateFormat(endDate),
	)
}

func (s *orderService) BatchOrderProcess(orders []models.Order) (models.BatchOrderResponse, error) {
	var batchOrderResponse models.BatchOrderResponse
	batchOrderResponse.Summary.TotalOrders = len(orders)

	for _, order := range orders {
		var processedOrder models.BatchProcessedOrder
		processedOrder.CustomerName = order.CustomerName
		validator := models.NewOrderValidator(order)
		if errMap := validator.Validate(); errMap != nil {
			processedOrder.Status = "rejected"
			processedOrder.Reason = "missing fields"
			batchOrderResponse.ProcessedOrders = append(batchOrderResponse.ProcessedOrders, processedOrder)
			continue
		}

		orderID, err := s.orderRepo.Insert(order)
		processedOrder.ID = orderID
		if err != nil {
			processedOrder.Status = "rejected"
			switch {
			case errors.Is(err, models.ErrForeignKeyConstraintOrderMenu):
				processedOrder.Reason = "menu item does not exist"
			case errors.Is(err, models.ErrNegativeQuantity):
				processedOrder.Reason = "insufficient inventory"
			default:
				processedOrder.Reason = "internal server error"
			}
			batchOrderResponse.ProcessedOrders = append(batchOrderResponse.ProcessedOrders, processedOrder)
			continue
		}

		totalOrderPrice, _ := s.orderRepo.GetBatchTotalOrderPrice(orderID)

		processedOrder.Status = "accepted"
		processedOrder.Total = totalOrderPrice
		batchOrderResponse.Summary.Accepted++
		batchOrderResponse.ProcessedOrders = append(batchOrderResponse.ProcessedOrders, processedOrder)
	}

	var orderIDs []int
	for _, order := range batchOrderResponse.ProcessedOrders {
		if order.Status == "accepted" {
			orderIDs = append(orderIDs, order.ID)
			batchOrderResponse.Summary.TotalRevenue += order.Total
		}
	}
	batchOrderResponse.Summary.Rejected = batchOrderResponse.Summary.TotalOrders - batchOrderResponse.Summary.Accepted
	inventoryUpdates, err := s.orderRepo.GetBatchInventoryUpdates(orderIDs)
	if err != nil {
		slog.Error("Failed to get inventory updates", "error", err)
	}
	batchOrderResponse.Summary.InventoryUpdates = append(batchOrderResponse.Summary.InventoryUpdates, inventoryUpdates...)
	return batchOrderResponse, err
}
