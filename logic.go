package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/HDIOES/anime-sheduler/dao"
	"github.com/nats-io/nats.go"
)

//InitEventHandler struct
type InitEventHandler struct {
	db             *sql.DB
	settings       *Settings
	natsConnection *nats.Conn
	client         *http.Client
}

func (ieh *InitEventHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if response, resErr := ieh.client.Get(ieh.settings.ShikimoriSheduleURL); resErr != nil {
		log.Println(resErr)
	} else {
		sheduleItems := make([]SheduleItem, 50)
		if decodeErr := json.NewDecoder(response.Body).Decode(&sheduleItems); decodeErr != nil {
			log.Println(decodeErr)
		} else {

		}
	}
}

//SheduleItem struct
type SheduleItem struct {
	NextEpisode   int64         `json:"next_episode"`
	NextEpisodeAt ShikimoriTime `json:"next_episode_at"`
	Duration      int64         `json:"duration"`
	Anime         Anime         `json:"anime"`
}

//Anime struct
type Anime struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Russian string `json:"russian"`
	Image   Image  `json:"image"`
	URL     string `json:"url"`
}

//Image struct
type Image struct {
	Original string `json:"original"`
	Preview  string `json:"preview"`
	X96      string `json:"x96"`
	X48      string `json:"x48"`
}

//ShikimoriTime struct
type ShikimoriTime struct {
	time.Time
}

//UnmarshalJSON unmarshales ShikimoriTime correctly
func (sts *ShikimoriTime) UnmarshalJSON(b []byte) (err error) {
	s := string(b)
	s = s[1 : len(s)-1]
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return err
	}
	sts.Time = t
	return nil
}

func (sts *ShikimoriTime) toDateValue() *string {
	value := sts.Format(time.RFC3339)
	return &value
}

//UpdateSheduleHandler struct
type UpdateSheduleHandler struct {
	db       *sql.DB
	settings *Settings
	adao     *dao.AnimeDAO
}

func (ush *UpdateSheduleHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

}
