package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"github.com/nats-io/nats.go"
)

const notificationType = "notificationType"

//InitEventHandler struct
type InitEventHandler struct {
	db             *sql.DB
	adao           *AnimeDAO
	sdao           *SubscriptionDAO
	settings       *Settings
	natsConnection *nats.Conn
}

func (ieh *InitEventHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logReqErr := logRequest(r)
	if logReqErr != nil {
		HandleError(logReqErr)
		return
	}
	animeDtos, userDtos, err := ieh.sdao.GetSubscriptionsAndMarkAnimesAsNotified()
	if err != nil {
		HandleError(err)
		return
	}
	countOfNotifications := len(animeDtos)
	notifications := make([]TelegramCommandMessage, 0, countOfNotifications)
	for i := 0; i < countOfNotifications; i++ {
		telegramID, parseErr := strconv.ParseInt(userDtos[i].ExternalID, 10, 64)
		if parseErr != nil {
			HandleError(parseErr)
			return
		}
		notification := TelegramCommandMessage{
			TelegramID: telegramID,
			Type:       notificationType,
			Text:       fmt.Sprintf("%s. Новый эпизод вышёл в эфир. Не пропустите!", animeDtos[i].EngName),
		}
		notifications = append(notifications, notification)
	}
	for _, notification := range notifications {
		if sendNotificationErr := ieh.sendNotification(notification); sendNotificationErr != nil {
			HandleError(sendNotificationErr)
			return
		}
	}
}

func (ieh *InitEventHandler) sendNotification(notification TelegramCommandMessage) error {
	data, dataErr := json.Marshal(notification)
	if dataErr != nil {
		return errors.WithStack(dataErr)
	}
	if publishErr := ieh.natsConnection.Publish(ieh.settings.NatsSubject, data); publishErr != nil {
		return errors.WithStack(publishErr)
	}
	return nil
}

//SheduleItem struct
type SheduleItem struct {
	NextEpisodeAt *ShikimoriTime `json:"next_episode_at"`
	Anime         Anime          `json:"anime"`
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
		return errors.WithStack(err)
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
	settings    *Settings
	adao        *AnimeDAO
	HTTPGateway *HTTPGateway
}

func (ush *UpdateSheduleHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logReqErr := logRequest(r)
	if logReqErr != nil {
		HandleError(logReqErr)
		return
	}
	if httpStatus, resReader, resErr := ush.HTTPGateway.Get(ush.settings.ShikimoriSheduleURL); resErr != nil {
		HandleError(errors.WithStack(resErr))
	} else {
		if httpStatus != 200 {
			HandleError(errors.New("Http status not equals 200"))
		}
		sheduleItems := []SheduleItem{}
		if decodeErr := json.NewDecoder(resReader).Decode(&sheduleItems); decodeErr != nil {
			HandleError(decodeErr)
		} else {
			if updateSheduleErr := ush.adao.UpdateAnimes(sheduleItems); updateSheduleErr != nil {
				HandleError(updateSheduleErr)
			}
		}
	}
}

//TelegramCommandMessage struct
type TelegramCommandMessage struct {
	Type string `json:"type"`
	//fields for notification and /start
	TelegramID int64  `json:"telegramId"`
	Text       string `json:"text"`
}
