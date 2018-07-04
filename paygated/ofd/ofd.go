package ofd

import (
	"github.com/aavzz/daemon/log"
	"github.com/aavzz/dqb-paygate/paygated/storage"
	"github.com/aavzz/dqb-paygate/paygated/billing"
        "time"
	"github.com/spf13/viper"
)

type ofd interface {
        init()
	RegisterReceipt(pid, cid, t, vat, phone, email string, sum float32) error
	ReceiptInfo(pid string) *ResponseOk
}

var Ofd ofd

//InitOfd initializes connection to fiscal data operator
func InitOfd() {
        switch viper.GetString("ofd.type") {
        case "ekam":
                Ofd = new(ekam)
        default:
                log.Error("Unknown OFD type: " + viper.GetString("ofd.type"))
        }

        if Ofd == nil {
                log.Fatal("Cannot proceed to initialize OFD")
        }

        if viper.GetString("ofd.url") == "" {
                log.Fatal("ofd.url cannot be empty")
        }

        if viper.GetString("ofd.token") == "" {
                log.Fatal("ofd.token cannot be empty")
        }

	Ofd.init()

	//Store a new receipt locally
       	go func() {
		for {
			time.Sleep(10 * time.Second)

			s := storage.Storage.GetUnhandledOfd()
			if s != nil {
				for k, v := range s {
       			               	r := storage.Storage.StoreReceipt(v.id)
				}
               		}
       		}
       	}()

	//Register pending receipt with OFD or update its status
       	go func() {
		for {
			time.Sleep(10 * time.Second)

			s := storage.Storage.GetPendingReceipts()
			if s != nil {
				for k, v := range s {
 	                      		ri := Ofd.ReceiptInfo(v.Uuid)
 	                      		if ri == nil {
						ui := billing.Billing.GetUserInfo(cid)
 	                      			err := Ofd.RegisterReceipt(v.Uuid, v.Cid, v.Type, v.Vat, ui.PhoneNumber, ui.Email, v.Sum)
 	                      			if err == nil {
						}
					} else {
						switch ri.Status {
						case "printed":
       			               			storage.Storage.ReceiptPrinted(v.Id)
       			              			// delete? XXX storage.Storage.SetHandledOfd(k)
							if viper.GetString("notification.url") == "" {
       	                         				storage.Storage.SetHandledNotification(k, "ofd")
							}
						case "error":
       			               			storage.Storage.ReceiptError(v.Id)
						}
					}
				}
			}
		}
	}


}

