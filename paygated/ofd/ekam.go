package ofd

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"errors"
	"bytes"
	"io/ioutil"
	"github.com/aavzz/daemon/log"
	"github.com/aavzz/dqb-paygate/paygated/billing"
	"github.com/spf13/viper"
)

type ekam struct {
	token, url string
}

//init initializes ekam
func (e *ekam) init() {
	e.token = viper.GetString("ofd.token")
	e.url = viper.GetString("ofd.url")
}

//RegisterReceipt sends receipt info to ekam
func (e *ekam) RegisterReceipt(pid, cid, t string, sum float32) error {

	var rcptLines ReceiptLines
	var rcpt ReceiptRequest

	ui := billing.Billing.GetUserInfo(cid)
	if ui != nil {
		if rcpt.Email != "" {
			rcpt.Email = ui.Email
		}
		if rcpt.PhoneNumber != "" {
  			rcpt.PhoneNumber = ui.PhoneNumber
		}
	}
	rcptLines.Price = sum
      	rcptLines.Quantity = 1
      	rcptLines.Title = "Услуги"
      	rcptLines.TotalPrice = sum
      	//rcptLines.VatRate    

  	rcpt.OrderId = pid
  	rcpt.OrderNumber = pid
  	rcpt.Type = t
  	rcpt.ShouldPrint = false
  	rcpt.CashAmount = 0
  	rcpt.ElectronAmount = sum
  	rcpt.CashierName = ""
  	rcpt.Draft = true
  	rcpt.Lines = append(rcpt.Lines, rcptLines)

	jsonValue, err := json.MarshalIndent(rcpt, "", "    ")
	if err != nil {
		log.Error("Ekam: " + err.Error())
		return err
	}

	log.Info(string(jsonValue))

	req, err := http.NewRequest("POST", e.url, bytes.NewBuffer(jsonValue))
	if err != nil {
		log.Error("Ekam: " + err.Error())
		return err
	}
	req.Header.Set("X-Access-Token", e.token)
	req.Header.Set("Content-Type", "application/json")

	c := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	resp, err := c.Do(req)
	if err != nil {
		log.Error("Ekam: " + err.Error())
		return err
	}
        if resp != nil {
                defer resp.Body.Close()

		switch resp.StatusCode {
                case 200:  
                case 201:  
                        body, err := ioutil.ReadAll(resp.Body)
                        if err != nil {
				log.Error("Ekam: " + err.Error())
                                return err
                        }
                        var v ResponseOk
                        if err := json.Unmarshal(body, &v); err != nil {
				log.Error("Ekam: " + err.Error())
                                return err
                        }
	jsonValue, _ := json.MarshalIndent(v, "", "    ") //XXX
	log.Info("200" + string(jsonValue))
			return nil
                case 422:  
                        body, err := ioutil.ReadAll(resp.Body)
                        if err != nil {
                                log.Error(err.Error())
                                return err
                        }
                        var v ResponseError
                        if err := json.Unmarshal(body, &v); err != nil {
				log.Error("Ekam: " + err.Error())
                                return err
                        }
	jsonValue, _ := json.MarshalIndent(v, "", "    ") //XXX
	log.Info("422" + string(jsonValue))
                        return errors.New(resp.Status)
		default:
			log.Error("Ekam: " + resp.Status)
                        return errors.New(resp.Status)
                }
        } else {
		log.Error("Ekam: no response from ekam")
                return errors.New("No response from ekam")
        }

        return nil
}


