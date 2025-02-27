package games

import (
	"time"

	"github.com/navazjm/pixelarcade/internal/webapp/utils/validator"
)

type Score struct {
	ID                 int64     `json:"id"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	Version            int32     `json:"version"`
	IsActive           bool      `json:"is_active"`
	GameID             int64     `json:"game_id"`
	UserID             int64     `json:"user_id"`
	Score              int64     `json:"score"`
	UserName           string    `json:"user_name"`            // derived from inner join w/ users_preferences table
	UserProfilePicture string    `json:"user_profile_picture"` // derived from inner join w/ users_preferences table
}

func ValidateScore(v *validator.Validator, score int64) {
	v.Check(score > 0, "score", "must be non negative number")
}
