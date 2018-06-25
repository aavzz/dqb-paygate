package ofd

import (
	"github.com/aavzz/gqb-paygate/paygated/storage"
        "github.com/aavzz/daemon/log"
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
		Ofd.init()
        }

        go func() {
            for {
                time.Sleep(10 * time.Second)

                s, err := storage.Storage.GetUnhandledOfd()
                if err != nil {
                        log.Fatal(err.Error())
                }
                for k, v := range s {
                        err = Ofd.RegisterReceipt(v.Cid, v.Type, v.Sum)
                        if err == nil {
                                err = storage.Storage.SetHandledOfd(k)
                                if err != nil {
                                        log.Fatal(err.Error())
                                }
                        }
                }
            }
        }()

        return nil
}

