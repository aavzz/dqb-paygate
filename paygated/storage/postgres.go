package storage

import (
	"database/sql"
	"github.com/aavzz/daemon/log"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"github.com/satori/go.uuid"
)


type postgres struct {
        dbh *sql.DB
}

//init connects to postgres DB
func (s *postgres) init() {
	dbh, err := sql.Open("postgres", "host="+viper.GetString("storage.host")+" user="+viper.GetString("storage.user")+
                   " password="+viper.GetString("storage.pass")+" dbname="+viper.GetString("storage.name")+" sslmode=disable")
	if err != nil {
		log.Fatal(err.Error())
	}
	if err = dbh.Ping(); err != nil {
		if err := dbh.Close(); err != nil {
			log.Fatal(err.Error())
		}
		log.Fatal(err.Error())
	}
	s.dbh = dbh
}

//StorePayment stores payment in local database and checks if it has really been stored
func (s *postgres) StorePayment(cpid,cid,channel,terminal,direction string, sum float32) *Payment {

        var p Payment

	//check if the payment already exists
        rows, err := s.dbh.Query("SELECT payment_id, tstamp_paygate FROM payments WHERE payment_channel=$1 AND channel_payment_id=$2", channel, cpid)
        if err != nil {
		log.Error("Postgres: " + err.Error())
                return nil
        }
        defer rows.Close()
        if !rows.Next() {
		uuId := uuid.NewV4()
		if cpid != "" {
	        	_, err := s.dbh.Exec("INSERT INTO payments(channel_payment_id, payment_sum, payment_subject_id, payment_channel, channel_terminal_id, payment_direction, payment_id) VALUES ($1, $2, $3, $4, $5, $6, $7)",
                                  cpid, sum, cid, channel, terminal, direction, uuId.String())
	       		if err != nil {
				log.Error("Postgres: " + err.Error())
				return nil
        		}
		} else {
	        	_, err := s.dbh.Exec("INSERT INTO payments(paymant_sum, payment_subject_id, payment_channel, channel_terminal_id, payment_direction, payment_id) VALUES ($1, $2, $3, $4, $5, $6)",
                                  sum, cid, channel, terminal, direction, uuId.String())
	       		if err != nil {
				log.Error("Postgres: " + err.Error())
				return nil
        		}
		}

		rows1, err := s.dbh.Query("SELECT id,tstamp_paygate FROM payments WHERE payment_channel=$1 AND channel_payment_id=$2", channel, cpid)
		if err != nil {
			log.Error("Postgres: " + err.Error())
			return nil
		}
		defer rows1.Close()
		if !rows1.Next() {
			log.Error("Postgres: Cannot find inserted payment " + channel + " " + cpid)
			return nil
		}
		if err := rows1.Scan(&p.LocalId,&p.Tstamp); err != nil {
			log.Error("Postgres: " + err.Error())
			return nil
        	}
        } else {
		if err := rows.Scan(&p.LocalId,&p.Tstamp); err != nil {
			log.Error("Postgres: " + err.Error())
			return nil
        	}
		log.Info("Postgres: Incoming payment has already been saved: " + channel + " " + cpid)
	}
        return &p
}

//GetUnhandledBilling gets unprocessed db records
func (s *postgres) GetUnhandledBilling() map[uint64]Unhandled {
	m := make(map[uint64]Unhandled)
	rows, err := s.dbh.Query("SELECT id, payment_subject_id, payment_sum, payment_id, payment_channel, payment_vat FROM payments WHERE tstamp_billing is null")
        if err != nil {
		log.Error("Postgres: " + err.Error())
            return nil
        }
        defer rows.Close()
        for rows.Next() {

            var id uint64
            var sum float32
            var channel,cid,pid,vat string
            if err := rows.Scan(&id,&cid,&sum,&pid,&channel,&vat); err != nil {
		log.Error("Postgres: " + err.Error())
                return nil
            }
		m[id] = Unhandled{
			Cid: cid,
			Sum: sum,
			PaymentId: pid,
			Vat: vat,
			Channel: channel,
		}
        }
        return m
}

//GetUnhandledOfd gets unprocessed db records
func (s *postgres) GetUnhandledOfd() map[uint64]Unhandled {
	m := make(map[uint64]Unhandled)
	rows,err := s.dbh.Query("SELECT id, payment_subject_id, payment_sum, payment_id, payment_channel, payment_direction, payment_vat FROM payments WHERE tstamp_ofd is null")
        if err != nil {
		log.Error("Postgres: " + err.Error())
            return nil
        }
        defer rows.Close()
        for rows.Next() {

            var id uint64
            var sum float32
            var channel,cid,pid,t,vat string
            if err := rows.Scan(&id,&cid,&sum,&pid,&channel,&t,&vat); err != nil {
		log.Error("Postgres: " + err.Error())
                return nil
            }
		m[id] = Unhandled{
			Cid: cid,
			Sum: sum,
			PaymentId: pid,
			Vat: vat,
			Channel: channel,
			Type: t,
		}
        }
        return m
}

//GetUnhandledNotifier gets unprocessed db records
func (s *postgres) GetUnhandledNotification() map[uint64]Unhandled {
	m := make(map[uint64]Unhandled)
	rows,err := s.dbh.Query("SELECT id, payment_subject_id, payment_sum, payment_id, payment_channel, payment_direction, payment_vat FROM payments WHERE tstamp_notification is null")
        if err != nil {
		log.Error("Postgres: " + err.Error())
            return nil
        }
        defer rows.Close()
        for rows.Next() {

            var id uint64
            var sum float32
            var channel,cid,pid,t,vat string
            if err := rows.Scan(&id,&cid,&sum,&pid,&channel,&t,&vat); err != nil {
		log.Error("Postgres: " + err.Error())
                return nil
            }
		m[id] = Unhandled{
			Cid: cid,
			Sum: sum,
			PaymentId: pid,
			Vat: vat,
			Channel: channel,
			Type: t,
		}
        }
        return m
}

//SetHandledBilling marks db record as processed
func (s *postgres) SetHandledBilling(id uint64) error {
        if _, err := s.dbh.Exec("UPDATE payments set tstamp_billing=current_timestamp where id=$1", id); err != nil {
		log.Error("Postgres: " + err.Error())
            return err
        }
	return nil
}

//SetHandledOfd marks db record as processed
func (s *postgres) SetHandledOfd(id uint64) error {
        if _, err := s.dbh.Exec("UPDATE payments set tstamp_ofd=current_timestamp where id=$1", id); err != nil {
		log.Error("Postgres: " + err.Error())
            return err
        }
	return nil
}

//SetHandledNotification marks db record as processed
func (s *postgres) SetHandledNotification(id uint64, addr string) error {
        if _, err := s.dbh.Exec("UPDATE payments set tstamp_notification=current_timestamp, notification_sent_to=$1 where id=$2", addr, id); err != nil {
		log.Error("Postgres: " + err.Error())
            return err
        }
	return nil
}

//Shutdown closes db connection                  
func (s *postgres) Shutdown() error {
        if s.dbh != nil {
                if err := s.dbh.Close(); err != nil {
                        log.Error("Postgres: " + err.Error())
                        return err
                }
        }             
        return nil
}

