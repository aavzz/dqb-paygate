package storage

import (
	"github.com/spf13/viper"
	"github.com/aavzz/daemon/log"
)

type storage interface {
        init()
	StorePayment(pid,cid,channel,terminal,direction string, sum float32) *Payment
	StoreReceipt(pid uint64, rid string) *Receipt
	GetUnhandledBilling() map[uint64]Unhandled
	GetUnhandledOfd() map[uint64]Unhandled
	GetUnhandledNotification() map[uint64]Unhandled
	GetPendingReceipts() map[uint64]Receipt
	SetHandledBilling(id uint64) error
	// XXX SetHandledOfd(id uint64) error
	SetHandledNotification(id uint64, addr string) error
        Shutdown() error
}

type Unhandled struct {
	id uint64
        Cid,PaymentId,Channel,Type,Vat string
        Sum float32
}

type Payment struct {
	LocalId, Tstamp string
}

type Receipt struct {
	Id uint64
	Cid, Uuid, Type, Vat string
        Sum float32
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

