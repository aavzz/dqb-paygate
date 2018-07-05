package ofd

import (
	"github.com/aavzz/daemon/log"
	"github.com/aavzz/dqb-paygate/paygated/storage"
        "time"
	"github.com/spf13/viper"
)

type ofd interface {
        init()
	RegisterReceipt(pid, cid, t, vat string, sum float32) error
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

       	go func() {
		for {
			time.Sleep(10 * time.Second)

			s := storage.Storage.GetUnhandledOfd()
			if s == nil {
				continue
			}
 	               	for k, v := range s {
				switch v.Type {
				case "in": 
					v.Type = "sale"
				case "out":
					v.Type = "return"
				}
				r := Ofs.ReceiptInfo(v.paymentId)
				if r == nil {
					Ofd.RegisterReceipt(v.PaymentId, v.Cid, v.Type, v.Vat, v.Sum)
				} else {
					switch r.Status {
					case "pending":
						continue
					case "printed":
						storage.Storage.SetHandledOfd(k)
						if viper.GetString("notification.url") == "" {
							storage.Storage.SetHandledNotification(k, "ofd")
						}
					case "error":
						Ofd.RegisterReceipt(v.PaymentId, v.Cid, v.Type, v.Vat, v.Sum)
					}
       	                	}
			}
       		}
       	}()
}

