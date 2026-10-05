package handler

import (
	"errors"
	"net/http"

	"go_marketplace_v1/internal/model"
	"go_marketplace_v1/internal/service"
)

type UserHandler struct {
	users *service.UserService
}

func NewUserHandler(users *service.UserService) *UserHandler {
	return &UserHandler{users: users}
}

type createUserRequest struct {
	Name string `json:"name"`
}

type userResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func userToResponse(user model.User) userResponse {
	return userResponse{
		ID:   user.ID,
		Name: user.Name,
	}
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	var request createUserRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}

	created, err := h.users.Create(r.Context(), request.Name)
	if err != nil {
		if errors.Is(err, service.ErrInvalidName) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}

		writeError(w, http.StatusInternalServerError, "внутренняя ошибка")
		return
	}

	_ = writeJSON(w, http.StatusCreated, userToResponse(created))
}
