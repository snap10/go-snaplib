package mgoutils

import (
	"log"
	"net/http"
	"time"

	"github.com/snap10/go-snaplib/httputils"
	"github.com/snap10/go-snaplib/logging"

	"gopkg.in/mgo.v2"
)

type Config struct {
	User     string
	Password string
	Host     []string
	session  *mgo.Session
}

var config Config

//Get a Session for MGO
func (config *Config) GetSession() *mgo.Session {
	user := config.User
	host := config.Host
	password := config.Password

	if config.session == nil {
		config.createDbSession(host, user, password)
	}
	return config.session
}
func (config *Config) createDbSession(hosts []string, user, password string) {
	logging.Info.Printf("Creating DB-Session to %s", hosts)
	var err error
	session, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    hosts,
		Username: user,
		Password: password,
		Timeout:  60 * time.Second,
	})
	config.session = session
	if err != nil {
		log.Fatalf("[createDbSession]: %s\n", err)
	}
}

func HandleMgoError(w http.ResponseWriter, err error) {
	if err == mgo.ErrNotFound {
		httputils.DisplayAppError(w, err, "Not found in Database", http.StatusNotFound)
	} else {
		logging.Error.Println(err.Error())
		httputils.DisplayAppError(w, err, "Unexpected Problem with Database", http.StatusInternalServerError)
	}
}
