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
	"strings"
	"strconv"
	"fmt"
)

// Handler calls the right function to send message via specified channel.
func Sdm(w http.ResponseWriter, r *http.Request) {

        cmd := r.FormValue("Command")
        payId := r.FormValue("PaymentID")
        terminal := r.FormValue("TerminalId")
        userId := r.FormValue("ClientId")
        sum := r.FormValue("Ammount")


        w.Header().Set("Content-type", "text/xml")

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
        if m, _ := regexp.MatchString(`^\d+,\d\d$`, sum); !m {
                    w.Write([]byte("wrong sum format"))
			log.Info("Pskb: wrong sum format")	
                    return
        }
        if m, _ := regexp.MatchString(viper.GetString("billing.uid_format"), userId); !m {
                    w.Write([]byte("wrong uid format"))
			log.Info("Pskb: wrong uid format")
                    return
        }

	sum = strings.Replace(sum, ",", ".", -1)

	value, _ := strconv.ParseFloat(sum, 32)
        sumFloat := float32(value)  

        switch cmd {
        case "check":
                ui := billing.Billing.GetUserInfo(userId)
                if ui != nil {
			w.Write([]byte("<?xml version=\"1.0\" encoding=\"windows-1251\"?>"))
			w.Write([]byte("<Response>"))
	        	w.Write([]byte("  <Result>0</Result>"))
	        	w.Write([]byte("  <PaymentId>" + payId + "</PaymentId>"))
	        	w.Write([]byte("  <Description>OK</Description>"))
	        	w.Write([]byte("</Response>"))
                } else {
			w.Write([]byte("<?xml version=\"1.0\" encoding=\"windows-1251\"?>"))
			w.Write([]byte("<Response>"))
	        	w.Write([]byte("  <Result>1</Result>"))
	        	w.Write([]byte("  <PaymentId>" + payId + "</PaymentId>"))
	        	w.Write([]byte("  <Description>USER NOT FOUND(" + userId + ")</Description>"))
	        	w.Write([]byte("</Response>"))
                }
        case "payment":
                p := storage.Storage.StorePayment(payId, userId, "sdm", terminal, "in", sumFloat)
                if p != nil {
			w.Write([]byte("<?xml version=\"1.0\" encoding=\"windows-1251\"?>"))
			w.Write([]byte("<Response>"))
	        	w.Write([]byte("  <Result>0</Result>"))
	        	w.Write([]byte("  <PaymentNumber>dqb" + fmt.Sprintf("%d", p.Number) + "</PaymentNumber>"))
	        	w.Write([]byte("  <PaymentId>" + payId + "</PaymentId>"))
	        	w.Write([]byte("  <PaymentTime>" + p.Time + "</PaymentTime>"))
	        	w.Write([]byte("  <Description>OK</Description>"))
	        	w.Write([]byte("</Response>"))
                } else {
			w.Write([]byte("<?xml version=\"1.0\" encoding=\"windows-1251\"?>"))
			w.Write([]byte("<Response>"))
	        	w.Write([]byte("  <Result>1</Result>"))
	        	w.Write([]byte("  <PaymentId>" + payId + "</PaymentId>"))
	        	w.Write([]byte("  <Description>DB FAILURE</Description>"))
	        	w.Write([]byte("</Response>"))
                }
        }
}

