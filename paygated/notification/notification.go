package notification

import (
	"github.com/spf13/viper"
	"github.com/aavzz/notifier"
	"github.com/aavzz/daemon/log"
	"github.com/aavzz/dqb-paygate/paygated/billing"
        "github.com/aavzz/dqb-paygate/paygated/storage"
        "github.com/aavzz/dqb-paygate/paygated/ofd"
	"strings"
	"time"
	"fmt"
	"io/ioutil"
)

type notification struct {
	emailTemplate, smsTemplate string
}

var nn notification

//InitNotification initializes notification
func InitNotification() {

	if viper.GetString("notification.url") != "" {

		if viper.GetString("notification.email_subject") == "" {
			log.Fatal("Notification: email_subject is not set")
		}
		if viper.GetString("notification.email_sender_name") == "" {
			log.Fatal("Notification: email_sender_name is not set")
		}
		if viper.GetString("notification.email_sender_address") == "" {
			log.Fatal("Notification: email_sender_address is not set")
		}
		if viper.GetString("notification.sms_template") == "" {
			log.Fatal("Notification: sms_template is not set")
		}
		if viper.GetString("notification.email_template") == "" {
			log.Fatal("Notification: email_template is not set")
		}
		t, err := ioutil.ReadFile(viper.GetString("notification.sms_template"))
		if err != nil {
			log.Fatal(err.Error())
		}
		nn.smsTemplate = string(t)

		t, err = ioutil.ReadFile(viper.GetString("notification.email_template"))
		if err != nil {
			log.Fatal(err.Error())
		}
		nn.emailTemplate = string(t)

	        go func() {
			for {
				time.Sleep(10 * time.Second)

				n := storage.Storage.GetUnhandledNotification()
				if n == nil {
					continue
				}

				for k, v := range n {
					r := ofd.Ofd.ReceiptInfo(v.PaymentId)
					if r == nil || r.FiscalData.RegistrationNumber == "" {
						continue
					}
					ui := billing.Billing.GetUserInfo(v.Cid)
					if ui == nil {
						log.Error("Failed to get user info: " + v.Cid)
						continue
					}
					if ui.Type == "ul" {
                               	         	storage.Storage.SetHandledNotification(k, "ul")
						continue
					}
					addr := ui.PhoneNumber
					channel := viper.GetString("notification.sms_channel")
					template := nn.smsTemplate
					if ui.Email != "" {
						addr = ui.Email
						channel = "email"
						template = nn.emailTemplate
					}
					if addr == "" || channel == "" {
               	         			storage.Storage.SetHandledNotification(k, "")
						continue
					}
					message := template

					t, err := time.Parse(time.RFC3339, r.CreatedAt)
					t = t.Local()
					if err != nil {
						log.Error("Failed to parse time")
					}
					year, month, day := t.Date()
					date := fmt.Sprintf("%02d.%02d.%d", day,month,year)
					hm := fmt.Sprintf("%02d:%02d", t.Hour(), t.Minute())
					tz, _ := t.Zone()

					receiptType := "Приход"
					if r.Type == "ReturnReceiptRequest" {
						receiptType = "Возврат прихода"
					}

					message = strings.Replace(message, "%CONTRACT%", v.Cid, -1)
					message = strings.Replace(message, "%DATE%", date, -1)
					message = strings.Replace(message, "%TIME%", hm, -1)
					message = strings.Replace(message, "%ZONE%", tz, -1)
					message = strings.Replace(message, "%SUM%", r.Amount, -1)
					message = strings.Replace(message, "%FPD%", r.FiscalData.Fpd, -1)
					message = strings.Replace(message, "%RECEIPT_TYPE%", receiptType, -1)
					message = strings.Replace(message, "%REG_KKT%", r.FiscalData.RegistrationNumber, -1)

					if channel == "email" {
						message = strings.Replace(message, "%EMAIL%", addr, -1)
						message = strings.Replace(message, "%SHIFT%", r.FiscalData.RetailShiftNumber, -1)
						message = strings.Replace(message, "%RECEIPT_NUM%", fmt.Sprintf("%d",r.FiscalData.ReceiptNumber), -1)
						message = strings.Replace(message, "%FD%", r.FiscalData.FdNumber, -1)
						message = strings.Replace(message, "%FN_NUM%", r.FiscalData.FactoryFnNumber, -1)
						message = strings.Replace(message, "%INN%", r.FiscalData.OrganizationInn, -1)
						message = strings.Replace(message, "%SENDER_EMAIL%", viper.GetString("notification.email_sender_address"), -1)
						err = notifier.NotifyEmail(viper.GetString("notification.url"), addr,
								 viper.GetString("notification.email_subject"),
								 viper.GetString("notification.email_sender_name"),
								 viper.GetString("notification.email_sender_address"),
								 message)
                           			if err == nil {
                               	         			storage.Storage.SetHandledNotification(k, addr)
               	           			} else {
							log.Info(err.Error())
						}
					} else {
						err := notifier.NotifySMS(viper.GetString("notification.url"), channel, "+" + addr, message)
               	 				if err == nil {
               	         				storage.Storage.SetHandledNotification(k, addr)
						} else {
							log.Info(err.Error())
						}
					}
                               	}
                        }
		}()
	} else {
		log.Info("Notification: url not set, notifying via OFD")
	}
}

