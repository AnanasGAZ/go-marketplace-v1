package model

type User struct {
	ID           int64
	Name         string
	Login        *string
	PasswordHash *string `json:"-"` //это правило для отправки данных наружу: когда Go превращает пользователя в JSON-ответ, поле PasswordHash пропускается. Хэш при этом остаётся в программе и базе — просто клиенту его не отправляют.
}

//указатель Login *string можно оставить nil
//SQL NULL прямо в поле string стандартная библиотека Go  вернёт ошибку
