package storage

import (
	"database/sql"
	"errors"
	"github.com/aavzz/daemon/log"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
)


type postgres struct {
        dbh *sql.DB
}

//init connects to postgres DB
func (s *postgres) init() error {
	dbh, err := sql.Open("postgres", "host="+viper.GetString("storage.host")+" user="+viper.GetString("storage.user")+
                   " password="+viper.GetString("storage.pass")+" dbname="+viper.GetString("storage.name")+" sslmode=disable")
	if err != nil {
		log.Fatal(err.Error())
	}
	if err = dbh.Ping(); err != nil {
		if err = dbh.Close(); err != nil {
			log.Fatal(err.Error())
		}
		log.Fatal(err.Error())
	}
	s.dbh = dbh
        return nil
}

//StorePayment stores payment in local database and checks if it has really been stored
func (s *postgres) StorePayment(pid,cid,channel,terminal string, sum float32) error {

        if _, err := s.dbh.Exec("INSERT INTO payments(channel_payment_id, paymant_sum, payment_subject_id, payment_channel, channel_terminal_id, tstamp_paygate) VALUES ($1, $2, $3, $4, $5, current_timestamp)",
                                  pid, sum, cid, channel, terminal); err != nil {
            return err
        }
        return nil

}

//GetUnhandledBilling gets unprocessed db records
func (s *postgres) GetUnhandledBilling() (map[uint64]Unhandled, error) {
	m := make(map[uint64]Unhandled)
	rows, err := s.dbh.Query("SELECT id,payment_subject_id,payment_sum,channel_payment_id,payment_channel FROM payments WHERE tstamp_billing is null")
        if err != nil {
            return nil, err
        }
        defer rows.Close()
        for rows.Next() {

            var id uint64
            var sum float32
            var channel,cid,pid string
            if err := rows.Scan(&id,&cid,&sum,&pid,&channel); err != nil {
                return nil, err
            }
		m[id] = Unhandled{
			Cid: cid,
			Sum: sum,
			Payment_id: pid,
			Channel: channel,
		}
        }

        return m, nil
}

//GetUnhandledOfd gets unprocessed db records
func (s *postgres) GetUnhandledOfd() (map[uint64]Unhandled, error) {
	m := make(map[uint64]Unhandled)
	rows,err := s.dbh.Query("SELECT id, payment_subject_id, payment_sum, channel_payment_id, payment_channel, payment_direction FROM payments WHERE tstamp_ofd is null")
        if err != nil {
            return nil, err
        }
        defer rows.Close()
        for rows.Next() {

            var id uint64
            var sum float32
            var channel,cid,pid,t string
            if err := rows.Scan(&id,&cid,&sum,&pid,&channel,&t); err != nil {
                return nil, err
            }
		m[id] = Unhandled{
			Cid: cid,
			Sum: sum,
			Payment_id: pid,
			Channel: channel,
			Type: t,
		}
        }

        return m, nil
}

//SetHandledBilling marks db record as processed
func (s *postgres) SetHandledBilling(id uint64) error {
        if _, err := s.dbh.Exec("UPDATE payments set tstamp_billing=current_timestamp where id=$1", id); err != nil {
            return err
        }
	return nil
}

//SetHandledOfd marks db record as processed
func (s *postgres) SetHandledOfd(id uint64) error {
        if _, err := s.dbh.Exec("UPDATE payments set tstamp_ofd=current_timestamp where id=$1", id); err != nil {
            return err
        }
	return nil
}

//Shutdown closes db connection                  
func (s *postgres) Shutdown() error {
        if s.dbh != nil {
                if err := s.dbh.Close(); err != nil {
                        log.Error(err.Error())
                        return err
                }
        }             
        return nil
}

