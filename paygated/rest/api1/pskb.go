/*
Package api1 implements version 1 of notifyd API.
*/
package api1

import (
	"github.com/aavzz/dqb-paygate/paygated/billing"
	"github.com/aavzz/d1b-paygate/paygated/storage"
	"github.com/spf13/viper"
	"net/http"
	"regexp"
	"strconv"
)

// Pskb processes payment requests from pskb.
func Pskb(w http.ResponseWriter, r *http.Request) {

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
	case "get_info":
                if _, err := billing.Billing.GetUserInfo(userId); err == nil {
                    w.Write([]byte("status=0"))
                } else {
                    w.Write([]byte("status=-1"))
                }
	case "payment":
                if err := storage.Storage.StorePayment(payId, userId, "pskb", terminal, sumFloat); err == nil {
                    w.Write([]byte("status=0"))
                } else {
                    w.Write([]byte("status=-3"))
                }
	}
}

