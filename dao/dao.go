package dao

import (
	"database/sql"
	"time"
)

//AnimeDAO struct
type AnimeDAO struct {
	Db *sql.DB
}

//AnimeDTO struct
type AnimeDTO struct {
	ID            int64
	ExternalID    string
	RusName       string
	EngName       string
	ImageURL      string
	NextEpisodeAt time.Time
}

func (adao *AnimeDAO) updateAnimes() {

}

//SubscriptionDAO struct
type SubscriptionDAO struct {
	Db *sql.DB
}

//SubcriptionDTO struct
type SubcriptionDTO struct {
	userID  int64
	animeID int64
}

//UserDAO struct
type UserDAO struct {
	Db *sql.DB
}

//UserDTO struct
type UserDTO struct {
	ID               int64
	ExternalID       string
	TelegramUsername string
}
