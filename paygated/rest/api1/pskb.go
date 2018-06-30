/*
Package api1 implements version 1 of notifyd API.
*/
package api1

import (
	"github.com/aavzz/daemon/log"
	"github.com/aavzz/dqb-paygate/paygated/billing"
	"github.com/aavzz/dqb-paygate/paygated/storage"
	"github.com/spf13/viper"
	"net/http"
	"regexp"
	"strconv"
)

// Pskb processes payment requests from pskb.
func Pskb(w http.ResponseWriter, r *http.Request) {

	login := r.FormValue("duser")
	pass := r.FormValue("dpass")
	if login != viper.GetString("pskb.login") || pass != viper.GetString("pskb.pass") {
		w.WriteHeader(403)
		log.Info("Pskb: Authentivation failed")
		return
	}

        w.Header().Set("Content-type", "text/html")

	userId := r.FormValue("cid")
        if m, _ := regexp.MatchString("^" + viper.GetString("billing.uid_format") + "$", userId); !m {
                    w.Write([]byte("wrong uid format"))
			log.Info("Pskb: wrong uid format")
                    return
        }

	cmd := r.FormValue("uact")
	switch cmd {
	case "get_info":
                ui := billing.Billing.GetUserInfo(userId)
                if ui != nil {
                    w.Write([]byte("status=0"))
                } else {
                    w.Write([]byte("status=-1"))
                }
	case "payment":
		payId := r.FormValue("trans")
		terminal := r.FormValue("term")
		sum := r.FormValue("sum")

	        if m, _ := regexp.MatchString(`^\d+$`, payId); !m {
                    w.Write([]byte("payment id is not numeric"))
			log.Info("Pskb: payment id is not numeric")
                    return
		}
		if m, _ := regexp.MatchString(`^\d+$`, terminal); !m {
                    w.Write([]byte("terminal is not numeric"))
			log.Info("Pskb: terminal is not numeric")
                    return
		}
		if m, _ := regexp.MatchString(`^\d+\.\d\d$`, sum); !m {
                    w.Write([]byte("wrong sum format"))
			log.Info("Pskb: wrong sum format")
                    return
		}
		value, _ := strconv.ParseFloat(sum, 32)
        	sumFloat := float32(value)  

                p := storage.Storage.StorePayment(payId, userId, "pskb", terminal, "in", sumFloat)
                if p != nil {
                    w.Write([]byte("status=0"))
                } else {
                    w.Write([]byte("status=-3"))
                }
	default:
                    w.Write([]byte("wrong command"))
			log.Info("Pskb: wrong command")
	}
}

