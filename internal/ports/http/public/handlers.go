package public

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"log/slog"
	"net/http"
	"strconv"
	"tz_effective/internal/entities"
	"tz_effective/internal/ports/http/public/utils"
)

// CreateSubscription создает новую запись о подписке
// @Summary Создание подписки
// @Description Создает новую запись о подписке пользователя
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param subscription body entities.Subscriptions true "Данные подписки"
// @Success 201 {object} map[string]int64 "id созданной подписки"
// @Failure 400 {string} string "Ошибка в запросе"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /subscriptions [post]
func (s *Server) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	var sub entities.Subscriptions
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := utils.ValidateUUID(sub.UserID); err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := utils.ValidateDate(sub.StartDate); err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	if sub.EndDate != nil {
		if err := utils.ValidateDate(*sub.EndDate); err != nil {
			RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	id, err := s.Service.CreateSubscription(r.Context(), &sub)
	if err != nil {
		slog.Error("Failed to create subscription", "error", err)
		RespondWithError(w, http.StatusInternalServerError, "failed to create subscription")
		return
	}
	RespondWithJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

// GetSubscription получает информацию о подписке по ID
// @Summary Получение подписки
// @Description Получает детальную информацию о подписке по её ID
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path int true "ID подписки"
// @Success 200 {object} entities.Subscriptions "Данные подписки"
// @Failure 400 {string} string "Некорректный ID"
// @Failure 404 {string} string "Подписка не найдена"
// @Router /subscriptions/{id} [get]
func (s *Server) GetSubscription(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid id")
		return
	}
	sub, err := s.Service.GetSubscription(r.Context(), id)
	if err != nil {
		RespondWithError(w, http.StatusNotFound, "subscription not found")
		return
	}
	RespondWithJSON(w, http.StatusOK, sub)
}

// UpdateSubscription обновляет информацию о подписке
// @Summary Обновление подписки
// @Description Обновляет информацию о существующей подписке
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path int true "ID подписки"
// @Param subscription body entities.Subscriptions true "Новые данные подписки"
// @Success 200 {object} map[string]string "Статус обновления"
// @Failure 400 {string} string "Ошибка в запросе"
// @Failure 404 {string} string "Подписка не найдена"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /subscriptions/{id} [put]
func (s *Server) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var sub entities.Subscriptions
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := utils.ValidateUUID(sub.UserID); err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := utils.ValidateDate(sub.StartDate); err != nil {
		RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	if sub.EndDate != nil {
		if err := utils.ValidateDate(*sub.EndDate); err != nil {
			RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
	}

	if err := s.Service.UpdateSubscription(r.Context(), id, &sub); err != nil {
		slog.Error("Failed to update subscription", "error", err)
		RespondWithError(w, http.StatusInternalServerError, "failed to update subscription")
		return
	}
	RespondWithJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// DeleteSubscription удаляет подписку по ID
// @Summary Удаление подписки
// @Description Удаляет существующую подписку по её ID
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path int true "ID подписки"
// @Success 204 {object} map[string]string "Статус удаления"
// @Failure 400 {string} string "Некорректный ID"
// @Failure 404 {string} string "Подписка не найдена"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /subscriptions/{id} [delete]
func (s *Server) DeleteSubscription(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := s.Service.DeleteSubscription(r.Context(), id); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "failed to delete subscription")
		return
	}
	RespondWithJSON(w, http.StatusNoContent, map[string]string{"status": "deleted"})
}

// ListSubscriptions возвращает список подписок с фильтрацией
// @Summary Список подписок
// @Description Получает список подписок с возможностью фильтрации
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param user_id query string false "ID пользователя (UUID)"
// @Param service_name query string false "Название сервиса"
// @Param start_date query string false "Дата начала подписки (MM-YYYY)"
// @Param end_date query string false "Дата окончания подписки (MM-YYYY)"
// @Success 200 {array} entities.Subscriptions "Список подписок"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /subscriptions [get]
func (s *Server) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	filter := entities.ListFilter{}
	if v := r.URL.Query().Get("user_id"); v != "" {
		filter.UserID = &v
	}
	if v := r.URL.Query().Get("service_name"); v != "" {
		filter.ServiceName = &v
	}
	if v := r.URL.Query().Get("start_date"); v != "" {
		filter.StartDate = &v
	}
	if v := r.URL.Query().Get("end_date"); v != "" {
		filter.EndDate = &v
	}

	subs, err := s.Service.ListSubscriptions(r.Context(), &filter)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "failed to list subscriptions")
		return
	}
	RespondWithJSON(w, http.StatusOK, subs)
}

// CalculateTotalCost рассчитывает суммарную стоимость подписок за период
// @Summary Расчет стоимости подписок
// @Description Рассчитывает суммарную стоимость всех подписок за выбранный период с фильтрацией
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param start_period query string true "Начало периода (MM-YYYY)"
// @Param end_period query string true "Конец периода (MM-YYYY)"
// @Param user_id query string false "ID пользователя (UUID)"
// @Param service_name query string false "Название сервиса"
// @Success 200 {object} entities.TotalCostResponse "Суммарная стоимость"
// @Failure 400 {string} string "Ошибка в параметрах запроса"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /subscriptions/cost [get]
func (s *Server) CalculateTotalCost(w http.ResponseWriter, r *http.Request) {
	startPeriod := r.URL.Query().Get("start_period")
	endPeriod := r.URL.Query().Get("end_period")

	if startPeriod == "" || endPeriod == "" {
		RespondWithError(w, http.StatusBadRequest, "start_period and end_period are required")
		return
	}

	if err := utils.ValidateDate(startPeriod); err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid start_period format: "+err.Error())
		return
	}
	if err := utils.ValidateDate(endPeriod); err != nil {
		RespondWithError(w, http.StatusBadRequest, "invalid end_period format: "+err.Error())
		return
	}

	filter := &entities.CostFilter{
		StartPeriod: startPeriod,
		EndPeriod:   endPeriod,
	}

	if userID := r.URL.Query().Get("user_id"); userID != "" {
		if err := utils.ValidateUUID(userID); err != nil {
			RespondWithError(w, http.StatusBadRequest, "invalid user_id format: "+err.Error())
			return
		}
		filter.UserID = &userID
	}

	if serviceName := r.URL.Query().Get("service_name"); serviceName != "" {
		filter.ServiceName = &serviceName
	}

	totalCost, err := s.Service.CalculateTotalCost(r.Context(), filter)
	if err != nil {
		slog.Error("Failed to calculate total cost", "error", err)
		RespondWithError(w, http.StatusInternalServerError, "failed to calculate total cost")
		return
	}

	response := entities.TotalCostResponse{
		TotalCost: totalCost,
	}

	RespondWithJSON(w, http.StatusOK, response)
}
