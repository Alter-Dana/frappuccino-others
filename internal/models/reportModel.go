package models

type ReportTotalSales struct {
	OrdersCompleted int     `json:"orders_completed"` // Number of completed orders
	TotalSales      float64 `json:"total_sales"`
}

type ReportPopularItem struct {
	Name           string  `json:"name"`
	Description    string  `json:"description"`
	Price          float64 `json:"price"`
	Rank           int     `json:"rank"`
	TotalItemsSold int     `json:"total_items_sold"`
}

type ReportMenuSearchItem struct {
	Id          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Relevance   float32 `json:"relevance"`
}

type ReportOrderSearchItem struct {
	Id           string   `json:"id"`
	CustomerName string   `json:"customer_name"`
	Items        []string `json:"items"`
	Total        float64  `json:"total"`
	Relevance    float32  `json:"relevance"`
}

type ReportSearch struct {
	MenuResults   []ReportMenuSearchItem  `json:"menu_items,omitempty"`
	OrdersResults []ReportOrderSearchItem `json:"orders,omitempty"`
	TotalMatches  int                     `json:"total_matches"`
}

type ReportOrderedItems struct {
	Period       string           `json:"period"`
	Month        string           `json:"month,omitempty"`
	Year         string           `json:"year,omitempty"`
	OrderedItems []map[string]int `json:"orderedItems"`
}
