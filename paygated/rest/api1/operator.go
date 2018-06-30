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
	"github.com/satori/go.uuid"
)

// Handler calls the right function to send message via specified channel.
func Operator(w http.ResponseWriter, r *http.Request) {

	uuid, err := uuid.NewV4()
	if err != nil {
                    w.Write([]byte(err.Error()))
		log.Error("Operator: " + err.Error())
		return
	}

        cmd := r.FormValue("uact")
        userId := r.FormValue("cid")
        sum := r.FormValue("sum")

        if m, _ := regexp.MatchString(`^\d+\.\d\d$`, sum); !m {
                    w.Write([]byte("wrong sum format"))
			log.Info("Operator: wrong sum format")
                    return
        }
        if m, _ := regexp.MatchString("^" + viper.GetString("billing.uid_format") + "$", userId); !m {
                    w.Write([]byte("wrong uid format"))
			log.Info("Operator: wrong uid format")
                    return
        }

	value, _ := strconv.ParseFloat(sum, 32)
	sumFloat := float32(value)

        w.Header().Set("Content-type", "text/json")

        switch cmd {
        case "receive":
                p := storage.Storage.StorePayment(uuid, userId, "operator", "billing", "in", sumFloat)
                if p != nil {
                    w.Write([]byte("OK"))
                } else {
                    w.Write([]byte("FAILURE"))
                }
        case "return":
                p := storage.Storage.StorePayment(uuid, userId, "operator", "billing", "out", sumFloat)
                if p != nil {
                    w.Write([]byte("OK"))
                } else {
                    w.Write([]byte("FAILURE"))
                }
        default:       
                    w.Write([]byte("wrong command"))   
                        log.Info("Pskb: wrong command")         
        }
}

