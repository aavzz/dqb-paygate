/*
Package rest implements REST interface of paygated.
*/
package rest

import (
	"context"
	"github.com/aavzz/daemon/log"
	"github.com/aavzz/daemon/pid"
	"github.com/aavzz/daemon/signal"
	"github.com/aavzz/dqb-paygate/paygated/rest/api1"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"net/http"
	"time"
)

// InitHttp sets up router.
func InitHttp() {
	r := mux.NewRouter()
	r.HandleFunc("/pskb.cgi", api1.Pskb).Methods("POST")
	r.HandleFunc("/pskb", api1.Pskb).Methods("POST")
	r.HandleFunc("/sdm.cgi", api1.Sdm).Methods("POST")
	r.HandleFunc("/sdm", api1.Sdm).Methods("POST")
	r.HandleFunc("/operator", api1.Operator).Methods("POST")

	s := &http.Server{
		Addr:     viper.GetString("address"),
		Handler:  r,
		ErrorLog: log.Logger("paygated"),
	}

	if viper.GetBool("daemonize") == true {
		signal.Quit(func() {
			ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
			log.Info("SIGQUIT received, exitting gracefully")
			s.Shutdown(ctx)
			pid.Remove()
		})
	}

	log.Info("Initializing http: " + viper.GetString("address")) //XXX

	if err := s.ListenAndServe(); err != nil {
		log.Fatal(err.Error())
	}
}

