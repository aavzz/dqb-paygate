package ofd

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/url"
	"errors"
	"bytes"
	"io/ioutil"
	"github.com/aavzz/daemon/log"
	"github.com/aavzz/dqb-paygate/paygated/billing"
	"github.com/spf13/viper"
	"strconv"
	"fmt"
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
func (e *ekam) RegisterReceipt(pid, cid, t, vat string, sum float32) error {

	var rcptLines ReceiptLines
	var rcpt ReceiptRequest

	ui := billing.Billing.GetUserInfo(cid)
	if viper.GetString("notification.url") == "" && ui != nil {
		rcpt.Email = ui.Email
		if rcpt.Email == "" {
  			rcpt.PhoneNumber = ui.PhoneNumber
		}
	}
	if rcpt.Email == "" {
		rcpt.Email = "nonexistent@nowhere.net"
	}
	rcptLines.Price = sum
      	rcptLines.Quantity = 1
      	rcptLines.Title = "Услуги по договору " + cid 
      	rcptLines.TotalPrice = sum
	if vat != "" {
		vatNum, err := strconv.ParseInt(vat, 10, 64)
		if err != nil {
			log.Error("Ekam: " + err.Error())
			return err
		}
      		rcptLines.VatRate = &vatNum
	}

  	rcpt.OrderId = pid
  	rcpt.OrderNumber = pid
  	rcpt.Type = t
  	rcpt.ShouldPrint = false
  	rcpt.CashAmount = 0
  	rcpt.ElectronAmount = sum
  	rcpt.CashierName = ""
	if viper.GetString("ofd.draft") == "false" {
  		rcpt.Draft = false
	} else {
  		rcpt.Draft = true
	}
  	rcpt.Lines = append(rcpt.Lines, rcptLines)

	jsonValue, err := json.Marshal(rcpt)
	if err != nil {
		log.Error("Ekam: " + err.Error())
		return err
	}

	if viper.GetString("ofd.verbose") == "true" {
		log.Info(string(jsonValue))
	}

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
                case 200, 201:  
			if viper.GetString("ofd.verbose") == "true" {
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
				jsonValue, _ := json.Marshal(v)
				log.Info("-200-" + string(jsonValue))
                        }
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
			jsonValue, _ := json.Marshal(v)
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

//ReceiptInfo sends receipt info to ekam
func (e *ekam) ReceiptInfo(pid string) *ResponseOk {

        req, err := http.NewRequest("GET", e.url, nil)
        if err != nil {
                log.Error("Ekam: " + err.Error())
                return nil
        }

	q := url.Values{}
	q.Add("order_id", pid)
	req.URL.RawQuery = q.Encode()

        req.Header.Set("X-Access-Token", e.token)
        req.Header.Set("Accept", "application/json")

        var v ResponseOkArray
        c := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
        resp, err := c.Do(req)
        if err != nil {
                log.Error("Ekam: " + err.Error())
                return nil
        }
        if resp != nil {

                defer resp.Body.Close()

                switch resp.StatusCode {
                case 200, 201:
                        body, err := ioutil.ReadAll(resp.Body)
                        if err != nil {          
                                log.Error("Ekam: " + err.Error())
                                return nil
                        }
                        if err := json.Unmarshal(body, &v); err != nil {
                                log.Error("Ekam: " + err.Error())
                                return nil
                        }
                        jsonValue, err := json.Marshal(v)
                        if err != nil {          
                                log.Error("Ekam: " + err.Error())
                                return nil
                        }
                        if viper.GetString("ofd.verbose") == "true" {
                                log.Info("200:" + fmt.Sprintf("%d", len(v.Items)) + string(jsonValue))
                        }         
			if len(v.Items) > 0 {
	                        return &v.Items[0]
			} else {
				return nil
			}
                case 422:                  
                        body, err := ioutil.ReadAll(resp.Body)
                        if err != nil {
                                log.Error(err.Error())
                                return nil
                        }
                        var v ResponseError
                        if err := json.Unmarshal(body, &v); err != nil {
                                log.Error("Ekam: " + err.Error())
                                return nil
                        }
                        jsonValue, _ := json.Marshal(v)
                        log.Info("422" + string(jsonValue))
                        return nil
                default:
                        log.Error("Ekam: " + resp.Status)
                        return nil
                }
        } else {
                log.Error("Ekam: no response from ekam")
                return nil
        }

        return nil
}
