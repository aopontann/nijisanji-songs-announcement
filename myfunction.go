package nsa

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog/log"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

func init() {
	// Register an HTTP function with the Functions Framework
	functions.HTTP("MyHTTPFunction", myHTTPFunction)
}

// Function myHTTPFunction is an HTTP handler
func myHTTPFunction(w http.ResponseWriter, r *http.Request) {
	// Your code here
	db, err := sql.Open("mysql", os.Getenv("DSN"))
	if err != nil {
		log.Fatal().Str("severity", "ERROR").Str("service", "sql.Open").Msg(err.Error())
	}

	err = CheckNewVideoTask(db)
	if err != nil {
		log.Fatal().Str("severity", "ERROR").Msg(err.Error())
	}

	if os.Getenv("ENV") != "dev" {
		err = MisskeyPostTask(db)
		if err != nil {
			log.Fatal().Str("severity", "ERROR").Msg(err.Error())
		}
		err = TweetTask(db)
		if err != nil {
			log.Fatal().Str("severity", "ERROR").Msg(err.Error())
		}
	} else {
		err = MailSendTask(db)
		if err != nil {
			log.Fatal().Str("severity", "ERROR").Msg(err.Error())
		}
	}

	// Send an HTTP response
	fmt.Fprintln(w, "OK")
}
