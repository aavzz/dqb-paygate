package billing

import (
        "database/sql"
	"errors"
	"github.com/aavzz/daemon/log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	"regexp"
)

type telix struct {
        dbh *sql.DB
}

//init connects to telix billing
func (b telix) init() {
        dbh, err := sql.Open("mysql", "host="+viper.GetString("billing.host")+" user="+viper.GetString("billing.user")+
                      " password="+viper.GetString("billing.pass")+" dbname="+viper.GetString("billing.name")+" sslmode=disable")
	if  err != nil {
                log.Fatal(err.Error())
        }
	log.Info("DB Ping") //XXX
        if err = dbh.Ping(); err != nil {
                if err = dbh.Close(); err != nil {
                        log.Fatal(err.Error())
                }
                log.Fatal(err.Error())
        }
	log.Info("DB Ping OK") //XXX
	b.dbh = dbh
}

//GetUserInfo checks if a given user exists
func (b telix) GetUserInfo(cid string) *UserInfo {
	rows, err := b.dbh.Query("SELECT COALESCE(phone, '') phone, COALESCE(mail, '') mail FROM contract WHERE cid=$1", cid)
        if err != nil {
		log.Error("Telix: " + err.Error() + ": " + cid)
		return nil
        }
	defer rows.Close()
        if !rows.Next() {
		log.Error("Telix: no user info found: " + cid)
		return nil
	}
	var ui UserInfo
        if err := rows.Scan(&ui.PhoneNumber, &ui.Email); err != nil {
	    log.Error("Telix: " + err.Error() + ": " + cid)
            return nil
        }

	//Normalize phone number (remove all non-digits)
        reg, err := regexp.Compile(`[^\d]`)
        if err != nil {
                log.Error("Telix: " + err.Error())
		return nil
        }
        ui.PhoneNumber = reg.ReplaceAllString(ui.PhoneNumber, "")

	if m, _ := regexp.MatchString(`^7\d\d\d\d\d\d\d\d\d\d$`, ui.PhoneNumber); !m {
		if ui.PhoneNumber != "" {
			log.Error("Junk phone number: " + ui.PhoneNumber + "(" + cid + ")");
			ui.PhoneNumber = "";
		} else {
			log.Error("Empty phone number: " + cid);
		}
        }

	if m, _ := regexp.MatchString(`^.+@.+\..+$`, ui.Email); !m {
		if ui.Email != "" {
			log.Error("Junk email: " + ui.Email + "(" + cid + ")");
			ui.Email = "";
		} else {
			log.Error("Empty email: " + cid);
		}
        }

	return &ui
}

//StorePayment stores payment and checks if it has really been stored
func (b telix) StorePayment(pid, cid, channel string, sum float32) error {

        //silently ignore double insertion attempts
	rows, err := b.dbh.Query("SELECT cid FROM payments WHERE agent=$1 AND trans=$2", channel, pid)
	if err != nil {
		log.Error("Telix: " + err.Error())
		return err
	}
	defer rows.Close()
        if rows.Next() {
            log.Info("Double insertion attempt, ignoring: " + channel + " " + pid)
            return nil
        }

	//insert payment
	t, err := b.dbh.Begin()
        if err != nil {
		log.Error("Telix: " + err.Error())
		return err
        }
	if _, err = t.Exec("INSERT INTO payments(trans, sum, cid, time, agent) VALUES ($1,$2,$3,current_timestamp,$4)", pid,sum,cid,channel); err != nil {
		if err := t.Rollback(); err != nil {
			log.Error("Telix: " + err.Error())
			return err
		}
		log.Error("Telix: " + err.Error())
		return err
	}
	if _, err = t.Exec("UPDATE contract SET balance=balance+$2 where cid=$1", cid,sum); err != nil {
		if err := t.Rollback(); err != nil {
			log.Error("Telix: " + err.Error())
			return err
		}
		log.Error("Telix: " + err.Error())
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
	rows1, err := b.dbh.Query("SELECT cid FROM payments WHERE agent=$1 AND trans=$2", channel, pid)
	if err != nil {
		log.Error("Telix: " + err.Error())
		return err
	}
	defer rows1.Close()
        if rows1.Next() {
            return nil
        }
	log.Error("Telix: Cannot find inserted payment:" + channel + " " + pid)
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

