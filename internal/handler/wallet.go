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
