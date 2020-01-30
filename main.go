package main

import (
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/lib/pq"
	"github.com/pkg/errors"

	"github.com/nats-io/nats.go"
	"go.uber.org/dig"
)

const (
	databaseURLEnvName               = "DATABASE_URL"
	maxOpenConnectionsEnvName        = "MAX_OPEN_CONNECTIONS"
	maxIdleConnectionsEnvName        = "MAX_IDLE_CONNECTIONS"
	databaseConnectionTimeoutEnvName = "DATABASE_CONNECTION_TIMEOUT"
	applicationPortEnvName           = "PORT"
	natsURLEnvName                   = "NATS_URL"
	natsSubjectEnvName               = "NATS_SUBJECT"
	shikimoriSheduleURLEnvName       = "SHIKIMORI_SHEDULE_URL"
)

func main() {
	container := dig.New()
	container.Provide(func() *Settings {
		if jsonFile, openErr := os.Open("settings.json"); openErr != nil {
			log.Panicln(openErr)
		} else {
			defer jsonFile.Close()
			decoder := json.NewDecoder(jsonFile)
			settings := &Settings{}
			if decodeErr := decoder.Decode(settings); decodeErr != nil {
				log.Panicln(decodeErr)
			} else {
				setSettingsFromEnv(settings)
				return settings
			}
		}
		panic("Unreachable code")
	})
	container.Provide(func(settings *Settings) (*sql.DB, *nats.Conn, *AnimeDAO, *SubscriptionDAO) {
		db, err := sql.Open("postgres", settings.DatabaseURL)
		if err != nil {
			log.Panicln(err)
		}
		db.SetMaxIdleConns(settings.MaxIdleConnections)
		db.SetMaxOpenConns(settings.MaxOpenConnections)
		timeout := strconv.Itoa(settings.ConnectionTimeout) + "s"
		timeoutDuration, durationErr := time.ParseDuration(timeout)
		if durationErr != nil {
			defer db.Close()
			log.Panicln(durationErr)
		} else {
			db.SetConnMaxLifetime(timeoutDuration)
		}
		natsConnection, ncErr := nats.Connect(settings.NatsURL)
		if ncErr != nil {
			log.Panicln(ncErr)
		}
		return db, natsConnection, &AnimeDAO{Db: db}, &SubscriptionDAO{Db: db}
	})
	container.Invoke(func(db *sql.DB, settings *Settings, natsConnection *nats.Conn) {
		mux := http.NewServeMux()
		adao := AnimeDAO{
			Db: db,
		}
		sdao := SubscriptionDAO{
			Db: db,
		}
		transport := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		mux.Handle("/updateShedule", &UpdateSheduleHandler{
			adao:     &adao,
			settings: settings,
			client: &http.Client{
				Transport: transport,
			},
		})
		mux.Handle("/initEvent", &InitEventHandler{
			sdao:           &sdao,
			settings:       settings,
			natsConnection: natsConnection,
		})
		log.Fatal(http.ListenAndServe(":"+strconv.Itoa(settings.ApplicationPort), mux))
	})
}

//Settings struct
type Settings struct {
	DatabaseURL         string `json:"databaseUrl"`
	MaxOpenConnections  int    `json:"maxOpenConnections"`
	MaxIdleConnections  int    `json:"maxIdleConnections"`
	ConnectionTimeout   int    `json:"connectionTimeout"`
	ApplicationPort     int    `json:"port"`
	NatsURL             string `json:"natsUrl"`
	NatsSubject         string `json:"natsSubject"`
	ShikimoriSheduleURL string `json:"shikimoriSheduleUrl"`
}

//StackTracer struct
type StackTracer interface {
	StackTrace() errors.StackTrace
}

//HandleError func
func HandleError(handledErr error) {
	if err, ok := handledErr.(StackTracer); ok {
		for _, f := range err.StackTrace() {
			log.Printf("%+s:%d\n", f, f)
		}
	} else {
		log.Println("Unknown error: ", err)
	}
}

func setSettingsFromEnv(settings *Settings) {
	if value := os.Getenv(databaseURLEnvName); value != "" {
		settings.DatabaseURL = value
	}
	if value := os.Getenv(maxOpenConnectionsEnvName); value != "" {
		if intValue, err := strconv.Atoi(value); err != nil {
			log.Panicln(err)
		} else {
			settings.MaxOpenConnections = intValue
		}
	}
	if value := os.Getenv(maxIdleConnectionsEnvName); value != "" {
		if intValue, err := strconv.Atoi(value); err != nil {
			log.Panicln(err)
		} else {
			settings.MaxIdleConnections = intValue
		}
	}
	if value := os.Getenv(databaseConnectionTimeoutEnvName); value != "" {
		if intValue, err := strconv.Atoi(value); err != nil {
			log.Panicln(err)
		} else {
			settings.ConnectionTimeout = intValue
		}
	}
	if value := os.Getenv(applicationPortEnvName); value != "" {
		if intValue, err := strconv.Atoi(value); err != nil {
			log.Panicln(err)
		} else {
			settings.ApplicationPort = intValue
		}
	}
	if value := os.Getenv(natsURLEnvName); value != "" {
		settings.NatsURL = value
	}
	if value := os.Getenv(natsSubjectEnvName); value != "" {
		settings.NatsSubject = value
	}
	if value := os.Getenv(shikimoriSheduleURLEnvName); value != "" {
		settings.ShikimoriSheduleURL = value
	}
}
