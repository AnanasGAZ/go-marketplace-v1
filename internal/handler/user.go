package handler

import (
	"errors"
	"net/http"
	"strconv"

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

	writeJSON(w, http.StatusCreated, userToResponse(created))
}
func (h *UserHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	usrIdStr := r.PathValue("user_id")
	id, err := strconv.ParseInt(usrIdStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "некорректный user_id")
		return
	}
	// -------------проверка прав ------------------
	currentID, ok := UserIDFromCtx(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "нет пользователя в контексте")
		return
	}
	if currentID != id {
		writeError(w, http.StatusForbidden, "доступ запрещён")
		return
	}
	//----------------------
	user, err := h.users.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "пользователь не найден")
			return
		}

		writeError(w, http.StatusInternalServerError, "внутренняя ошибка")
		return
	}
	writeJSON(w, http.StatusOK, userToResponse(user))
}
func (h *UserHandler) Me(w http.ResponseWriter, r *http.Request) {
	userId, ok := UserIDFromCtx(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "нет пользователя в контексте")
		return
	}

	user, err := h.users.GetByID(r.Context(), userId)
	if err != nil {
		if errors.Is(err, service.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "пользователь не найден")
			return
		}
		writeError(w, http.StatusInternalServerError, "внутренняя ошибка")
		return
	}

	writeJSON(w, http.StatusOK, userToResponse(user))
}
