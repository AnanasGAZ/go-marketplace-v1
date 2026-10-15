package handler

import (
	"errors"
	"net/http"

	"go_marketplace_v1/internal/model"
	"go_marketplace_v1/internal/service"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

type loginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type loginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

type registerRequest struct {
	Name     string `json:"name"`
	Login    string `json:"login"`
	Password string `json:"password"`
}

type registerResponse struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Login string `json:"login"`
}

func registrationToResponse(user model.User) registerResponse {
	login := ""
	if user.Login != nil {
		login = *user.Login
	}

	return registerResponse{
		ID:    user.ID,
		Name:  user.Name,
		Login: login,
	}
}

// w — куда обработчик записывает ответ
// r — входящий HTTP-запрос
// Login — метод структуры AuthHandler
// принимает запрос Login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var request loginRequest
	//корректный JSON?
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}
	//тут выдаем токен
	//h.auth.Login(...) ищет пользователя, проверяет пароль и создаёт токен;
	token, err := h.auth.Login(r.Context(), request.Login, request.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, service.ErrInvalidCredentials.Error())
			return
		}

		writeError(w, http.StatusInternalServerError, "внутренняя ошибка")
		return
	}
	//ответ
	//передавать токен в заголовке Authorization: Bearer <JWT>
	_ = writeJSON(w, http.StatusOK, loginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
	})
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var request registerRequest
	if err := decodeJSON(r, &request); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}

	created, err := h.auth.Register(r.Context(), request.Name, request.Login, request.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidName), errors.Is(err, service.ErrInvalidRegistration):
			writeError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrLoginAlreadyExists):
			writeError(w, http.StatusConflict, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "внутренняя ошибка")
		}
		return
	}

	_ = writeJSON(w, http.StatusCreated, registrationToResponse(created))
}
