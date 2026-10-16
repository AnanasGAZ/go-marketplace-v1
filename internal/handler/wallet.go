package handler

import (
	"errors"
	"net/http"
	"strconv"

	"go_marketplace_v1/internal/service"
)

type WalletHandler struct {
	wallets *service.WalletService
}

func NewWalletHandler(wallets *service.WalletService) *WalletHandler {
	return &WalletHandler{wallets: wallets}
}

type walletResponse struct {
	UserID  int64 `json:"user_id"`
	Balance int64 `json:"balance"`
}

func (h *WalletHandler) GetByUserID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("user_id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "некорректный user_id")
		return
	}

	// Проверка владельца: смотреть кошелёк может только сам пользователь.
	currentID, ok := UserIDFromCtx(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "нет пользователя в контексте")
		return
	}
	if currentID != id {
		writeError(w, http.StatusForbidden, "доступ запрещён")
		return
	}

	wallet, err := h.wallets.GetByUserID(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserDoesNotExist),
			errors.Is(err, service.ErrWalletNotFound):
			writeError(w, http.StatusNotFound, "кошелёк не найден")
		default:
			writeError(w, http.StatusInternalServerError, "внутренняя ошибка")
		}
		return
	}

	writeJSON(w, http.StatusOK, walletResponse{
		UserID:  wallet.UserID,
		Balance: wallet.Balance,
	})
}

type topUpRequest struct {
	Amount int64 `json:"amount"`
}

type topUpResponse struct {
	UserID        int64 `json:"user_id"`
	Balance       int64 `json:"balance"`
	TransactionID int64 `json:"transaction_id"`
}

func (h *WalletHandler) TopUp(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("user_id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "некорректный user_id")
		return
	}

	// Проверка владельца
	currentID, ok := UserIDFromCtx(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "нет пользователя в контексте")
		return
	}
	if currentID != id {
		writeError(w, http.StatusForbidden, "доступ запрещён")
		return
	}

	var request topUpRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}

	wallet, transactionID, err := h.wallets.TopUp(r.Context(), id, request.Amount)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidAmount):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrBalanceOverflow):
			writeError(w, http.StatusConflict, err.Error())
		case errors.Is(err, service.ErrUserDoesNotExist),
			errors.Is(err, service.ErrWalletNotFound):
			writeError(w, http.StatusNotFound, "не найдено")
		default:
			writeError(w, http.StatusInternalServerError, "внутренняя ошибка")
		}
		return
	}

	writeJSON(w, http.StatusOK, topUpResponse{
		UserID:        wallet.UserID,
		Balance:       wallet.Balance,
		TransactionID: transactionID,
	})
}
