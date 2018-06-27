/*
Package api1 implements version 1 of notifyd API.
*/
package api1

import (
	"github.com/aavzz/daemon/log"
	"github.com/aavzz/dqb-paygate/paygated/storage"
	"github.com/spf13/viper"
	"net/http"
	"regexp"
	"strconv"
)

// Handler calls the right function to send message via specified channel.
func Operator(w http.ResponseWriter, r *http.Request) {

        cmd := r.FormValue("uact")
        userId := r.FormValue("cid")
        sum := r.FormValue("sum")

        w.Header().Set("Content-type", "text/json")

        if m, _ := regexp.MatchString(`^\d+\.\d\d$`, sum); !m {
                    w.Write([]byte("wrong sum format"))
			log.Info("Operator: wrong sum format")
                    return
        }
        if m, _ := regexp.MatchString("^" + viper.GetString("billing.uid_format") + "$", userId); !m {
                    w.Write([]byte("wrong uid format"))
                    return
        }

	value, _ := strconv.ParseFloat(sum, 32)
	sumFloat := float32(value)

        switch cmd {
        case "receive":
                if err := storage.Storage.StorePayment("", userId, "operator", "billing", "in", sumFloat); err == nil {
                    w.Write([]byte("OK"))
                } else {
                    w.Write([]byte("FAILURE"))
                }
        case "return":
                if err := storage.Storage.StorePayment("", userId, "operator", "billing", "out", sumFloat); err == nil {
                    w.Write([]byte("OK"))
                } else {
                    w.Write([]byte("FAILURE"))
                }
        }
}

