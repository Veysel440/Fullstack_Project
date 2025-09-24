package domain

import "time"

type Item struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Price     float64   `json:"price"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateItemDTO struct {
	Name  string  `json:"name"  validate:"required,min=1,max=100"`
	Price float64 `json:"price" validate:"required,gte=0,lte=100000"`
}
