/*
Package api1 implements version 1 of notifyd API.
*/
package api1

import (
	"github.com/aavzz/dqb-paygate/paygated/billing"
	"github.com/aavzz/dqb-paygate/paygated/storage"
	"github.com/spf13/viper"
	"net/http"
	"regexp"
	"strconv"
)

// Handler calls the right function to send message via specified channel.
func Operator(w http.ResponseWriter, r *http.Request) {

        cmd := r.FormValue("uact")
        payId := r.FormValue("trans")
        terminal := r.FormValue("term")
        userId := r.FormValue("cid")
        sum := r.FormValue("sum")

        w.Header().Set("Content-type", "text/html")

        if m, _ := regexp.MatchString(`^\d+$`, payId); !m {
                    w.Write([]byte("payment id is not numeric"))
                    return
        }
        if m, _ := regexp.MatchString(`^\d+$`, terminal); !m {
                    w.Write([]byte("terminal is not numeric"))
                    return
        }
        if m, _ := regexp.MatchString(`^\d+\.\d\d$`, sum); !m {
                    w.Write([]byte("wrong sum format"))
                    return
        }
        if m, _ := regexp.MatchString(viper.GetString("billing.uid_format"), userId); !m {
                    w.Write([]byte("wrong uid format"))
                    return
        }

	value, _ := strconv.ParseFloat(sum, 32)
	sumFloat := float32(value)

        switch cmd {
        case "payment":
                if err := storage.Storage.StorePayment(payId, userId, "operator", terminal, sumFloat); err == nil {
                    w.Write([]byte("status=0"))
                } else {
                    w.Write([]byte("status=-3"))
                }
        }

}

