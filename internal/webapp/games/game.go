package games

import "time"

type Game struct {
	ID          int64     `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Version     int32     `json:"version"`
	IsActive    bool      `json:"is_active"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Logo        string    `json:"logo"`
	Src         string    `json:"src"`
	Controls    string    `json:"controls"`
	HasScore    bool      `json:"has_score"`
}

// TODO: Implement validation check during PostGameHandler, UpdateGameByIDHandler
