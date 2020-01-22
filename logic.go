package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/nats-io/nats.go"
)

//InitEventHandler struct
type InitEventHandler struct {
	db             *sql.DB
	adao           *AnimeDAO
	sdao           *SubscriptionDAO
	settings       *Settings
	natsConnection *nats.Conn
}

func (ieh *InitEventHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	animeDtos, userDtos, err := ieh.sdao.GetSubscriptions()
	if err != nil {
		log.Println(err)
		return
	}
	countOfNotifications := len(animeDtos)
	notifications := make([]Notification, 0, countOfNotifications)
	for i := 0; i < countOfNotifications; i++ {
		telegramID, parseErr := strconv.ParseInt(userDtos[i].ExternalID, 10, 64)
		if parseErr != nil {
			log.Println(parseErr)
			return
		}
		notification := Notification{
			TelegramID: telegramID,
			Type:       "notification",
			Text:       fmt.Sprintf("Аниме под названием %s вышло сейчас в эфир. Не пропустите!", animeDtos[i].EngName),
		}
		notifications = append(notifications, notification)
	}
	for _, notification := range notifications {
		if sendNotificationErr := ieh.sendNotification(notification); sendNotificationErr != nil {
			log.Println(sendNotificationErr)
			return
		}
	}
}

func (ieh *InitEventHandler) sendNotification(notification Notification) error {
	data, dataErr := json.Marshal(notification)
	if dataErr != nil {
		return dataErr
	}
	if publishErr := ieh.natsConnection.Publish(ieh.settings.NatsSubject, data); publishErr != nil {
		return publishErr
	}
	return nil
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
	adao     *AnimeDAO
	client   *http.Client
}

func (ush *UpdateSheduleHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if response, resErr := ush.client.Get(ush.settings.ShikimoriSheduleURL); resErr != nil {
		log.Println(resErr)
	} else {
		sheduleItems := make([]SheduleItem, 50)
		if decodeErr := json.NewDecoder(response.Body).Decode(&sheduleItems); decodeErr != nil {
			log.Println(decodeErr)
		} else {
			if updateSheduleErr := ush.adao.UpdateAnimes(sheduleItems); updateSheduleErr != nil {
				log.Println(updateSheduleErr)
			}
		}
	}
}

//Notification struct
type Notification struct {
	TelegramID int64      `json:"telegramId"`
	Type       string     `json:"type"`
	Text       string     `json:"text"`
	Animes     []AnimeDTO `json:"animes"`
	WebhookURL string     `json:"webhookUrl"`
}
