package main

import (
	"database/sql"
	"strconv"
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

//UpdateAnimes func
func (adao *AnimeDAO) UpdateAnimes(items []SheduleItem) error {
	tx, txErr := adao.Db.Begin()
	if txErr != nil {
		return txErr
	}
	for _, sheduleItem := range items {
		animeDTO, findErr := adao.findByExternalID(tx, strconv.FormatInt(sheduleItem.Anime.ID, 10))
		if findErr != nil {
			return rollbackByError(tx, findErr)
		}
		if animeDTO != nil {
			return rollbackByError(tx, adao.updateNextEpisodeAt(tx, animeDTO.ID, animeDTO.NextEpisodeAt))
		}
		//insert new anime
		if createErr := adao.create(tx,
			strconv.FormatInt(sheduleItem.Anime.ID, 10),
			sheduleItem.Anime.Russian,
			sheduleItem.Anime.Name,
			sheduleItem.Anime.URL,
			sheduleItem.NextEpisodeAt.Time); createErr != nil {
			return rollbackByError(tx, createErr)
		}
	}
	animeDtos, allAnimesErr := adao.allAnimes(tx)
	if allAnimesErr != nil {
		return rollbackByError(tx, allAnimesErr)
	}
	for _, animeDTO := range animeDtos {
		animeID, _ := strconv.ParseInt(animeDTO.ExternalID, 10, 64)
		found := false
		for _, sheduleItem := range items {
			if animeID == sheduleItem.Anime.ID {
				found = true
				break
			}
		}
		if !found {
			if deleteErr := adao.delete(tx, animeDTO.ID); deleteErr != nil {
				return rollbackByError(tx, deleteErr)
			}
		}
	}
	if commitErr := tx.Commit(); commitErr != nil {
		return commitErr
	}
	return nil
}

func rollbackByError(tx *sql.Tx, err error) error {
	if rollbackErr := tx.Rollback(); rollbackErr != nil {
		return rollbackErr
	}
	return err
}

func (adao *AnimeDAO) delete(tx *sql.Tx, id int64) error {
	return nil
}

func (adao *AnimeDAO) allAnimes(tx *sql.Tx) ([]AnimeDTO, error) {
	return nil, nil
}

func (adao *AnimeDAO) updateNextEpisodeAt(tx *sql.Tx, id int64, newNextEpisodeAt time.Time) error {
	return nil
}

func (adao *AnimeDAO) create(
	tx *sql.Tx,
	externalID string,
	rusName string,
	engName string,
	imageURL string,
	nextEpisodeAt time.Time) error {
	return nil
}

func (adao *AnimeDAO) findByExternalID(tx *sql.Tx, externalID string) (*AnimeDTO, error) {
	return nil, nil
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

//GetSubscriptions func
func (sdao *SubscriptionDAO) GetSubscriptions() ([]AnimeDTO, []UserDTO, error) {
	return nil, nil, nil
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
