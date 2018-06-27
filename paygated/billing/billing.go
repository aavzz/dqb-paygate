package billing

import (
	"github.com/spf13/viper"
	"github.com/aavzz/dqb-paygate/paygated/storage"
	"github.com/aavzz/daemon/log"
        "time"
)

type billing interface {
	init() error
	GetUserInfo(cid string) *UserInfo
	StorePayment(pid,cid,channel string, sum float32) error
	Shutdown() error
}

type UserInfo struct{
	Email, PhoneNumber string
}

var Billing billing

//InitBilling initializes connection to a billing
func InitBilling() error {
	switch viper.GetString("billing.type") {
        case "telix":
		Billing = new(telix)
	default:
		log.Error("Unknown billing type: " + viper.GetString("billing.type"))
        }

	if Billing != nil {
		Billing.init()

        	go func() {
          	  for {
                	time.Sleep(10 * time.Second)

		        s := storage.Storage.GetUnhandledBilling()
			if s != nil {
				for k, v := range s { 
                			err := Billing.StorePayment(v.Payment_id, v.Cid, v.Channel, v.Sum)
        				if err == nil {
						storage.Storage.SetHandledBilling(k)
					}
				}
			}
          	  }
		}()
	}

        return nil
}

