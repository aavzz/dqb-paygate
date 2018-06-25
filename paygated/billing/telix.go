package billing

import (
        "database/sql"
	"errors"
	"github.com/aavzz/daemon/log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
)

type telix struct {
        dbh *sql.DB
}

//init connects to telix billing
func (b telix) init() error {
        dbh, err := sql.Open("mysql", "host="+viper.GetString("billing.host")+" user="+viper.GetString("billing.user")+
                      " password="+viper.GetString("billing.pass")+" dbname="+viper.GetString("billing.name")+" sslmode=disable")
	if  err != nil {
                log.Fatal(err.Error())
        }
        if err = dbh.Ping(); err != nil {
                if err = dbh.Close(); err != nil {
                        log.Fatal(err.Error())
                }
                log.Fatal(err.Error())
        }
	b.dbh = dbh
        return nil
}

//GetUserInfo checks if a given user exists
func (b telix) GetUserInfo(cid string) (*UserInfo, error) {
	rows, err := b.dbh.Query("SELECT phone,mail,cid FROM contract WHERE cid=$1", cid)
        if err != nil {
		return nil, err
        }
        if !rows.Next() {
		return nil, errors.New("No rows found:" + cid)
	}

	var ui UserInfo
        if err := rows.Scan(&ui.PhoneNumber, &ui.Email); err != nil {
            return nil, err
        }
	return &ui, nil
}

//StorePayment stores payment and checks if it has really been stored
func (b telix) StorePayment(pid, cid, channel string, sum float32) error {

        //silently ignore double insertion attempts
	rows, err := b.dbh.Query("SELECT cid FROM payments WHERE agent=$1 AND trans=$2", channel, pid)
	if err != nil {
		return err
	}
        if rows.Next() {
            log.Info("Double insertion attempt, ignoring: " + pid)
            return nil
        }

	rows, err = b.dbh.Query("SELECT cid FROM payments WHERE agent=$1 AND trans=$2", channel, pid)
	if err != nil {
		return err
	}
        if !rows.Next() {
		return errors.New("Cannot find inserted payment " + pid)
	}

	t, err := b.dbh.Begin()
        if err != nil {
			return err
        }

	if _, err = t.Exec("INSERT INTO payments(trans, sum, cid, time, agent) VALUES ($1,$2,$3,current_timestamp,$4)", pid,sum,cid,channel); err != nil {
		if err := t.Rollback(); err != nil {
			return err
		}
		return err
	}

	if _, err = t.Exec("UPDATE contract SET balance=balance+$2 where cid=$1", cid,sum); err != nil {
		if err := t.Rollback(); err != nil {
			return err
		}
		return err
	}

	if _, err = t.Exec("UPDATE contract SET active=1 where cid=$1 AND balance>0 AND (active!=2 and active!=3 and active!=10)", cid); err != nil {
		if err := t.Rollback(); err != nil {
			return err
		}
		return err
	}

	if err = t.Commit(); err != nil {
		return err
	}

        //check the payment, just in case
	rows, err = b.dbh.Query("SELECT cid FROM payments WHERE agent=$1 AND trans=$2", channel, pid)
	if err != nil {
		return err
	}
        if rows.Next() {
            return nil
        }
	return errors.New("Cannot find inserted payment: " + pid)
}

//Shutdown closes billing connection
func (b telix) Shutdown() error {
        if b.dbh != nil {    
                if err := b.dbh.Close(); err != nil {
                        log.Error(err.Error())
			return err
                }
        }              
	return nil
}

