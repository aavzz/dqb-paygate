package billing

import (
        "database/sql"
	"errors"
	"github.com/aavzz/daemon/log"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
	"regexp"
	"strings"
)

type telix struct {
        dbh *sql.DB
}

//init connects to telix billing
func (b *telix) init() {
        dbh, err := sql.Open("mysql", viper.GetString("billing.user") + ":" + viper.GetString("billing.pass") + "@tcp(" + viper.GetString("billing.host") +
                      ":3306)" + "/" + viper.GetString("billing.name"))
	if  err != nil {
                log.Fatal(err.Error())
        }
        if err = dbh.Ping(); err != nil {
                if err := dbh.Close(); err != nil {
                        log.Fatal(err.Error())
                }
                log.Fatal(err.Error())
        }
	b.dbh = dbh
}

//GetUserInfo checks if a given user exists
func (b *telix) GetUserInfo(cid string) *UserInfo {

	rows, err := b.dbh.Query("SELECT COALESCE(phone, '') phone, COALESCE(mail, '') mail, subject FROM contract WHERE cid=?", cid)
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
	var subj uint8
	var email, phone string
        if err := rows.Scan(&phone, &email, &subj); err != nil {
	    log.Error("Telix: " + err.Error() + ": " + cid)
            return nil
        }

	switch subj {
	case 1:
		ui.Type="fl"
	case 2:
		ui.Type="ul"
	}

	arr := strings.Split(phone, ",")
	for _, i := range arr {

		//Normalize phone number (remove all non-digits)
        	reg, err := regexp.Compile(`[^\d]`)
        	if err != nil {
        	        log.Error("Telix: " + err.Error())
			return nil
        	}
        	i = reg.ReplaceAllString(i, "")

		if m, _ := regexp.MatchString(`^7\d\d\d\d\d\d\d\d\d\d$`, i); !m {
			if i != "" {
				log.Error("Junk phone number: " + i + "(" + cid + ")");
				i = "";
			} else {
				log.Error("Empty phone number: " + cid);
			}
        	} else {
			ui.PhoneNumber = ui.PhoneNumber + ",+" + i
		}

	}
       	reg, err := regexp.Compile(`,+$`)
       	if err != nil {
       	        log.Error("Telix: " + err.Error())
		return nil
       	}
       	ui.PhoneNumber = reg.ReplaceAllString(ui.PhoneNumber, "")


	arr := strings.Split(email, ",")
	for _, i := range arr {
		if m, _ := regexp.MatchString(`^.+@.+\..+$`, i); !m {
			if i != "" {
				log.Error("Junk email: " + i + "(" + cid + ")");
				ui.Email = "";
			} else {
				log.Error("Empty email: " + cid);
			}
        	} else {
			ui.Email = ui.Email + ",+" + i
		}
	}
       	reg, err := regexp.Compile(`,+$`)
       	if err != nil {
       	        log.Error("Telix: " + err.Error())
		return nil
       	}
       	ui.Email = reg.ReplaceAllString(ui.Email, "")

	log.Info("UserData: " + ui.Email + " " + ui.PhoneNumber) //XXX

	return &ui
}

//StorePayment stores payment and checks if it has really been stored
func (b *telix) StorePayment(pid, cid, channel, dir string, sum float32) error {

        //silently ignore double insertion attempts
	rows, err := b.dbh.Query("SELECT cid FROM payments WHERE agent=? AND trans=?", channel, pid)
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
	var result sql.Result
	if dir == "in" {
		result, err = t.Exec("INSERT INTO payments(trans, sum, cid, time, agent) VALUES (?,?,?,current_timestamp,?)", pid,sum,cid,channel)
	} else {
		result, err = t.Exec("INSERT INTO payments(trans, sum, cid, time, agent) VALUES (?,?,?,current_timestamp,?)", pid,sum * -1 ,cid,channel)
	}
	if err != nil {
		if err := t.Rollback(); err != nil {
			log.Error("Telix: " + err.Error())
			return err
		}
		log.Error("Telix: " + err.Error())
		return err
	}
	ra, err := result.RowsAffected()
	if err != nil {
		if err := t.Rollback(); err != nil {
			log.Error("Telix: " + err.Error())
			return err
		}
		log.Error("Telix: " + err.Error())
		return err
	}
	if ra != 1 {
		log.Error("Telix: insert into payment failed, rolling back")
		if err := t.Rollback(); err != nil {
			log.Error("Telix: " + err.Error())
			return err
		}
		return errors.New("Had to rollback")
	}
	if dir == "in" {
		result, err = t.Exec("UPDATE contract SET balance=balance+? where cid=?", sum, cid)
	} else {
		result, err = t.Exec("UPDATE contract SET balance=balance+? where cid=?", sum * -1, cid)
	}
	if err != nil {
		if err := t.Rollback(); err != nil {
			log.Error("Telix: " + err.Error())
			return err
		}
		log.Error("Telix: " + err.Error())
		return err
	}
	ra, err = result.RowsAffected()
	if err != nil {
		if err := t.Rollback(); err != nil {
			log.Error("Telix: " + err.Error())
			return err
		}
		log.Error("Telix: " + err.Error())
		return err
	}
	if ra != 1 {
		log.Error("Telix: update contract failed, rolling back")
		if err := t.Rollback(); err != nil {
			log.Error("Telix: " + err.Error())
			return err
		}
		return errors.New("Had to rollback")
	}
	if dir == "in" {
		_, err = t.Exec("UPDATE contract SET active=1 where cid=? AND balance>0 AND (active!=2 and active!=3 and active!=10)", cid)
	} else {
		_, err = t.Exec("UPDATE contract SET active=0 where cid=? AND balance<=0 AND (active!=2 and active!=3 and active!=10)", cid)
	}
	if err != nil {
		if err := t.Rollback(); err != nil {
			log.Error("Telix: " + err.Error())
			return err
		}
		log.Error("Telix: " + err.Error())
		return err
	}
	if err = t.Commit(); err != nil {
		log.Error("Telix: " + err.Error())
		return err
	}

        //check the payment, just in case
	rows1, err := b.dbh.Query("SELECT cid FROM payments WHERE agent=? AND trans=?", channel, pid)
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
func (b *telix) Shutdown() error {
        if b.dbh != nil {    
                if err := b.dbh.Close(); err != nil {
                        log.Error(err.Error())
			return err
                }
        }              
	return nil
}

