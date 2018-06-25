package billing

import (
	"github.com/spf13/viper"
	"github.com/aavzz/dqb-paygate/paygated/storage"
	"github.com/aavzz/daemon/log"
        "time"
)

type billing interface {
	init() error
	GetUserInfo(cid string) (*UserInfo, error)
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
		Billing.init()
        }

        go func() {
            for {
                time.Sleep(10 * time.Second)

	        s, err := storage.Storage.GetUnhandledBilling()
        	if err != nil {
                	log.Fatal(err.Error())
        	}
		for k, v := range s { 
                	err = Billing.StorePayment(v.Payment_id, v.Cid, v.Channel, v.Sum)
        		if err == nil {
				err = storage.Storage.SetHandledBilling(k)
        			if err != nil {
                			log.Fatal(err.Error())
        			}
			}
		}
            }
        }()

        return nil
}

