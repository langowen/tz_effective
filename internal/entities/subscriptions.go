package entities

type Subscriptions struct {
	ServiceName string  `json:"service_name"`
	Price       int64   `json:"price"`
	UserID      string  `json:"user_id"`
	StartDate   string  `json:"start_date"`
	EndDate     *string `json:"end_date,omitempty"`
}

type ListFilter struct {
	UserID      *string
	ServiceName *string
	StartDate   *string
	EndDate     *string
}

// CostFilter содержит параметры для фильтрации при подсчете стоимости подписок
type CostFilter struct {
	UserID      *string `json:"user_id,omitempty"`      // Фильтр по ID пользователя
	ServiceName *string `json:"service_name,omitempty"` // Фильтр по названию сервиса
	StartPeriod string  `json:"start_period"`           // Начало периода в формате MM-YYYY
	EndPeriod   string  `json:"end_period"`             // Конец периода в формате MM-YYYY
}

// TotalCostResponse структура для ответа с суммарной стоимостью
type TotalCostResponse struct {
	TotalCost int64 `json:"total_cost"` // Суммарная стоимость в рублях
}
