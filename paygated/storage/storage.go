package storage

import (
	"github.com/spf13/viper"
	"github.com/aavzz/daemon/log"
)

type storage interface {
        init()
	StorePayment(pid,cid,channel,terminal,direction string, sum float32) *Payment
	GetUnhandledBilling() map[uint64]Unhandled
	GetUnhandledOfd() map[uint64]Unhandled
	SetHandledBilling(id uint64) error
	SetHandledOfd(id uint64) error
	SetHandledNotification(id uint64) error
        Shutdown() error
}

type Unhandled struct {
        Cid,PaymentId,Channel,Type,Vat string
        Sum float32
}

type Payment struct {
	Number uint64
	Time string
}

var Storage storage

//InitStorage initializes storage
func InitStorage() {
        switch viper.GetString("storage.type") {
        case "postgres":
                Storage = new(postgres)
        default:
                log.Error("Unknown storage type: " + viper.GetString("storage.type"))
        }

        if Storage == nil {
                log.Fatal("Cannot proceed to initialize storage")                    
        }

	Storage.init()

}

