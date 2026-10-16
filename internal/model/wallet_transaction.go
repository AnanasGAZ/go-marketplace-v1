package model

import "time"

type WalletTransaction struct {
	ID        int64
	UserID    int64
	Amount    int64
	Kind      string
	CreatedAt time.Time
}
