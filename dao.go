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
	if insertNewAnimesErr := adao.insertNewAnimes(tx, items); insertNewAnimesErr != nil {
		return handleErr(tx, insertNewAnimesErr)
	}
	if deleteOldAnimersErr := adao.deleteOldAnimes(tx, items); deleteOldAnimersErr != nil {
		return handleErr(tx, deleteOldAnimersErr)
	}
	if commitErr := tx.Commit(); commitErr != nil {
		return handleErr(tx, commitErr)
	}
	return nil
}

func handleErr(tx *sql.Tx, err error) error {
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return rollbackErr
		}
	}
	return err
}

func (adao *AnimeDAO) insertNewAnimes(tx *sql.Tx, items []SheduleItem) error {
	//prepare findByExternalStmt
	findByExternalStmt, err := tx.Prepare("SELECT ID, EXTERNALID, RUSNAME, ENGNAME, IMAGEURL, NEXT_EPISODE_AT FROM ANIMES WHERE EXTERNALID = $1")
	if err != nil {
		return err
	}
	defer findByExternalStmt.Close()

	//prepare updateNextEpisodeAtStmt
	updateNextEpisodeAtStmt, err := tx.Prepare("UPDATE ANIMES SET NEXT_EPISODE_AT = $2 WHERE ID = $1")
	if err != nil {
		return err
	}
	defer updateNextEpisodeAtStmt.Close()

	//prepare createStmt
	createStmt, err := tx.Prepare("INSERT INTO ANIMES (EXTERNALID, RUSNAME, ENGNAME, IMAGEURL, NEXT_EPISODE_AT) VALUES($1, $2, $3, $4, $5)")
	if err != nil {
		return err
	}
	defer createStmt.Close()

	for _, sheduleItem := range items {
		animeDTO, findErr := adao.findByExternalID(findByExternalStmt, strconv.FormatInt(sheduleItem.Anime.ID, 10))
		if findErr != nil {
			return findErr
		}
		if animeDTO != nil {
			if updateErr := adao.updateNextEpisodeAt(updateNextEpisodeAtStmt, animeDTO.ID, animeDTO.NextEpisodeAt); updateErr != nil {
				return updateErr
			}
			continue
		}
		//insert new anime
		if createErr := adao.create(createStmt,
			strconv.FormatInt(sheduleItem.Anime.ID, 10),
			sheduleItem.Anime.Russian,
			sheduleItem.Anime.Name,
			sheduleItem.Anime.URL,
			sheduleItem.NextEpisodeAt.Time); createErr != nil {
			return createErr
		}
	}
	return nil
}

func (adao *AnimeDAO) deleteOldAnimes(tx *sql.Tx, items []SheduleItem) error {
	//prepare deleteStmt
	deleteStmt, err := tx.Prepare("DELETE FROM ANIMES WHERE ID = $1")
	if err != nil {
		return err
	}
	defer deleteStmt.Close()

	animeDtos, allAnimesErr := adao.allAnimes(tx)
	if allAnimesErr != nil {
		return allAnimesErr
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
			if deleteErr := adao.delete(deleteStmt, animeDTO.ID); deleteErr != nil {
				return deleteErr
			}
		}
	}
	return nil
}

func (adao *AnimeDAO) delete(stmt *sql.Stmt, id int64) error {
	if _, err := stmt.Exec(id); err != nil {
		return err
	}
	return nil
}

func (adao *AnimeDAO) allAnimes(tx *sql.Tx) ([]AnimeDTO, error) {
	result, resErr := tx.Query("SELECT ID, EXTERNALID, RUSNAME, ENGNAME, IMAGEURL, NEXT_EPISODE_AT FROM ANIMES")
	if resErr != nil {
		return nil, resErr
	}
	defer result.Close()
	animeDTOs := make([]AnimeDTO, 0, 50)
	for result.Next() {
		anime, err := adao.scanAsAnime(result)
		if err != nil {
			return nil, err
		}
		animeDTOs = append(animeDTOs, *anime)
	}
	return animeDTOs, nil
}

func (adao *AnimeDAO) updateNextEpisodeAt(stmt *sql.Stmt, id int64, newNextEpisodeAt time.Time) error {
	_, err := stmt.Exec(id, newNextEpisodeAt)
	if err != nil {
		return err
	}
	return nil
}

func (adao *AnimeDAO) create(
	stmt *sql.Stmt,
	externalID string,
	rusName string,
	engName string,
	imageURL string,
	nextEpisodeAt time.Time) error {
	_, err := stmt.Exec(externalID, rusName, engName, imageURL, nextEpisodeAt)
	if err != nil {
		return err
	}
	return nil
}

func (adao *AnimeDAO) scanAsAnime(result *sql.Rows) (*AnimeDTO, error) {
	var ID *sql.NullInt64
	var externalID *sql.NullString
	var rusname *sql.NullString
	var engname *sql.NullString
	var imageURL *sql.NullString
	var nextEpisodeAt *PqTime
	scanErr := result.Scan(&ID, &externalID, &rusname, &engname, &imageURL, &nextEpisodeAt)
	if scanErr != nil {
		return nil, scanErr
	}
	animeDTO := AnimeDTO{}
	if ID.Valid {
		animeDTO.ID = ID.Int64
	}
	if externalID.Valid {
		animeDTO.ExternalID = externalID.String
	}
	if rusname.Valid {
		animeDTO.RusName = rusname.String
	}
	if engname.Valid {
		animeDTO.EngName = engname.String
	}
	if imageURL.Valid {
		animeDTO.ImageURL = imageURL.String
	}
	if nextEpisodeAt.Valid {
		animeDTO.NextEpisodeAt = nextEpisodeAt.Time
	}
	return &animeDTO, nil
}

func (adao *AnimeDAO) findByExternalID(stmt *sql.Stmt, externalID string) (*AnimeDTO, error) {
	rows, err := stmt.Query(externalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if rows.Next() {
		anime, err := adao.scanAsAnime(rows)
		if err != nil {
			return nil, err
		}
		return anime, nil
	}
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
	rows, err := sdao.Db.Query("SELECT A.ID, A.EXTERNALID, A.RUSNAME, A.ENGNAME, A.IMAGEURL, A.NEXT_EPISODE_AT, TU.ID, TU.TELEGRAM_USER_ID, TU.TELEGRAM_USERNAME FROM ANIMES A JOIN SUBSCRIPTIONS S ON (A.ID = S.ANIME_ID AND A.NEXT_EPISODE_AT <= $1) JOIN TELEGRAM_USERS TU ON (TU.ID = S.TELEGRAM_USER_ID)", time.Now())
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	animes := make([]AnimeDTO, 0, 50)
	users := make([]UserDTO, 0, 50)
	for rows.Next() {
		anime, user, err := sdao.scanAsAnimeAndUser(rows)
		if err != nil {
			return nil, nil, err
		}
		animes = append(animes, *anime)
		users = append(users, *user)
	}
	return animes, users, nil
}

func (sdao *SubscriptionDAO) scanAsAnimeAndUser(rows *sql.Rows) (*AnimeDTO, *UserDTO, error) {
	//anime attributes
	var ID *sql.NullInt64
	var externalID *sql.NullString
	var rusname *sql.NullString
	var engname *sql.NullString
	var imageURL *sql.NullString
	var nextEpisodeAt *PqTime
	//user attributes
	var userID *sql.NullInt64
	var userExternalID *sql.NullString
	var username *sql.NullString
	scanErr := rows.Scan(&ID, &externalID, &rusname, &engname, &imageURL, &nextEpisodeAt, &userID, &userExternalID, &username)
	if scanErr != nil {
		return nil, nil, scanErr
	}
	animeDTO := AnimeDTO{}
	if ID.Valid {
		animeDTO.ID = ID.Int64
	}
	if externalID.Valid {
		animeDTO.ExternalID = externalID.String
	}
	if rusname.Valid {
		animeDTO.RusName = rusname.String
	}
	if engname.Valid {
		animeDTO.EngName = engname.String
	}
	if imageURL.Valid {
		animeDTO.ImageURL = imageURL.String
	}
	if nextEpisodeAt.Valid {
		animeDTO.NextEpisodeAt = nextEpisodeAt.Time
	}
	userDTO := UserDTO{}
	if userID.Valid {
		userDTO.ID = userID.Int64
	}
	if userExternalID.Valid {
		userDTO.ExternalID = userExternalID.String
	}
	if username.Valid {
		userDTO.TelegramUsername = username.String
	}
	return &animeDTO, &userDTO, nil
}

//UserDTO struct
type UserDTO struct {
	ID               int64
	ExternalID       string
	TelegramUsername string
}

//PqTime struct
type PqTime struct {
	Time  time.Time
	Valid bool
}

//Scan func
func (pt *PqTime) Scan(value interface{}) error {
	if value == nil {
		pt.Valid = false
		return nil
	}
	pt.Time = value.(time.Time)
	return nil
}
