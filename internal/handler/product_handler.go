package handler

import (
	"errors"
	"net/http"
	"strconv"

	"go_marketplace_v1/internal/model"
	"go_marketplace_v1/internal/repository"
	"go_marketplace_v1/internal/service"
)

type ProductHandler struct {
	svc *service.ProductService
}

func NewProductHandler(svc *service.ProductService) *ProductHandler {
	return &ProductHandler{svc: svc}
}

// --- Запросы и Ответы ---

type createProductRequest struct {
	Name  string `json:"name"`
	Price int64  `json:"price"`
	Stock int64  `json:"stock"`
}

type updateProductRequest struct {
	Name  *string `json:"name"` // Указатель, чтобы отличать "не передано" от "передана пустая строка"
	Price *int64  `json:"price"`
}

type setStockRequest struct {
	Stock int64 `json:"stock"`
}

type addStockRequest struct {
	Quantity int64 `json:"quantity"`
}

// --- Методы хэндлера ---

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createProductRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}

	p, err := h.svc.Create(r.Context(), req.Name, req.Price, req.Stock)
	if err != nil {
		// Проверяем ошибки валидации из сервиса
		if errors.Is(err, service.ErrInvalidName) ||
			errors.Is(err, service.ErrInvalidPrice) ||
			errors.Is(err, service.ErrInvalidStock) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "внутренняя ошибка")
		return
	}

	writeJSON(w, http.StatusCreated, p)
}

func (h *ProductHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")

	limit := 10 // значение по умолчанию
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	offset := 0 // значение по умолчанию
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	products, err := h.svc.GetAll(r.Context(), q, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "внутренняя ошибка")
		return
	}

	// Если товаров нет, вернем пустой массив [], а не null
	if products == nil {
		products = []model.Product{}
	}

	writeJSON(w, http.StatusOK, products)
}

func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("product_id") // Go 1.22+
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "некорректный ID товара")
		return
	}

	p, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			writeError(w, http.StatusNotFound, "товар не найден")
			return
		}
		writeError(w, http.StatusInternalServerError, "внутренняя ошибка")
		return
	}

	writeJSON(w, http.StatusOK, p)
}

func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("product_id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "некорректный ID товара")
		return
	}

	var req updateProductRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}

	p, err := h.svc.Update(r.Context(), id, req.Name, req.Price)
	if err != nil {
		if errors.Is(err, service.ErrEmptyUpdate) {
			writeError(w, http.StatusBadRequest, "должно быть передано хотя бы одно поле: name или price")
			return
		}
		if errors.Is(err, service.ErrInvalidName) || errors.Is(err, service.ErrInvalidPrice) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, repository.ErrProductNotFound) {
			writeError(w, http.StatusNotFound, "товар не найден")
			return
		}
		writeError(w, http.StatusInternalServerError, "внутренняя ошибка")
		return
	}

	writeJSON(w, http.StatusOK, p)
}

func (h *ProductHandler) SetStock(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("product_id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "некорректный ID товара")
		return
	}

	var req setStockRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}

	p, err := h.svc.SetStock(r.Context(), id, req.Stock)
	if err != nil {
		if errors.Is(err, service.ErrInvalidStock) {
			writeError(w, http.StatusBadRequest, "остаток не может быть отрицательным")
			return
		}
		if errors.Is(err, repository.ErrProductNotFound) {
			writeError(w, http.StatusNotFound, "товар не найден")
			return
		}
		writeError(w, http.StatusInternalServerError, "внутренняя ошибка")
		return
	}

	writeJSON(w, http.StatusOK, p)
}

func (h *ProductHandler) AddStock(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("product_id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "некорректный ID товара")
		return
	}

	var req addStockRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}

	p, err := h.svc.AddStock(r.Context(), id, req.Quantity)
	if err != nil {
		if errors.Is(err, service.ErrInvalidQuantity) {
			writeError(w, http.StatusBadRequest, "количество должно быть больше 0")
			return
		}
		if errors.Is(err, repository.ErrProductNotFound) {
			writeError(w, http.StatusNotFound, "товар не найден")
			return
		}
		if errors.Is(err, repository.ErrStockOverflow) {
			writeError(w, http.StatusConflict, "переполнение количества")
			return
		}
		writeError(w, http.StatusInternalServerError, "внутренняя ошибка")
		return
	}

	writeJSON(w, http.StatusOK, p)
}
