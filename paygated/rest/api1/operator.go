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

	login := r.FormValue("duser")
        pass := r.FormValue("dpass")
        if login != viper.GetString("operator.login") || pass != viper.GetString("operator.pass") {
                w.WriteHeader(403)
                log.Info("Operator: Authentivation failed")
                return
        }

	uuId := uuid.NewV4()

        cmd := r.FormValue("uact")
        userId := r.FormValue("cid")
        sum := r.FormValue("sum")
        agent := r.FormValue("agent")

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
        if sumFloat < 0.01 {
                    w.Write([]byte("wrong sum"))
			log.Info("Operator: wrong sum")
                    return
        }

        w.Header().Set("Content-type", "text/json")

        switch cmd {
        case "receive":
                p := storage.Storage.StorePayment(uuId.String(), userId, "operator", agent, "in", sumFloat)
                if p != nil {
                    w.Write([]byte("OK"))
                } else {
                    w.Write([]byte("FAILURE"))
                }
        case "return":
                p := storage.Storage.StorePayment(uuId.String(), userId, "operator", "billing", "out", sumFloat)
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

