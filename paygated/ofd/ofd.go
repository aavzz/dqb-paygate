package ofd

import (
	"github.com/aavzz/dqb-paygate/paygated/storage"
        "time"
	"github.com/spf13/viper"
)

type ofd interface {
        init() error
	RegisterReceipt(cid,t string, sum float32) error
}

var Ofd ofd


//InitOfd initializes connection to fiscal data operator
func InitOfd() error {
        switch viper.GetString("ofd.type") {
        case "ekam":
                Ofd = new(ekam)
        default:
                log.Error("Unknown OFD type: " + viper.GetString("ofd.type"))
        }

	if Ofd != nil {
		Ofd.init()

        	go func() {
            		for {
            		    time.Sleep(10 * time.Second)

 		               s := storage.Storage.GetUnhandledOfd()
 		               if s != nil {
 		               	for k, v := range s {
 		                      	 err := Ofd.RegisterReceipt(v.Cid, v.Type, v.Sum)
 		                      	 if err == nil {
 	      	                	         storage.Storage.SetHandledOfd(k)
 	      	                	 }
				}
                		}
       	     		}
        	}()
	}

        return nil
}

