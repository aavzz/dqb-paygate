/*
Package fiscal provides GO API to fiscal data operators
*/
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

type ReceiptLines struct {
      Price       float32
      Quantity    int
      Title       string
      Total_price float32
      Vat_rate    *int8
}

type ReceiptRequest struct {
//  Order_id        string
//  Order_number    string
  Type            string
  Email           string
  Phone_number    string
  Should_print    bool
  Cash_amount     float32
  Electron_amount float32
//  Cashier_name    string
  Draft           bool
  Lines           []ReceiptLines
}


//Must be exportable (used for EKAM response)
type ResponseOkLines struct {
      Id uint64
      Title string
      Quantity float32
      Total_price float32
      Price float32
      Vat_rate int
      Vat_amount float32
}

type ResponseOkFiscalData struct {
    Receipt_number uint64
    Model_number string
    Factory_kkt_number string
    Factory_fn_number string
    Registration_number string
    Fn_expired_period uint
    Fd_number uint
    Fpd uint
    Tax_system string
    Organisation_name string
    Organisation_inn string
    Address string
    Retail_shift_number string
    Ofd_name string
    Printed_at string
    Registration_date string
    Fn_expired_at string
}

type ResponseOk struct {
  id uint64
  uuid string
  t string `json:"type"`
  status string
  kkt_receipt_id uint
  amount float32
  cash_amount float32
  electron_amount float32
  lines []ResponseOkLines
  cashier_name string
  cashier_role string
  cashier_inn string
  transaction_address string
  email string
  phone_number string
  should_print bool
  order_id string
  order_number string
  created_at string
  updated_at string
  kkt_receipt_exists bool
  draft bool
  copy bool
  fiscal_data ResponseOkFiscalData
  receipt_url string
  online_cashier_url string
  error string
}

type ResponseError struct {
	Error string
}

//init initializes ekam
func (e *ekam) init() {
	e.token = viper.GetString("ofd.token")
	e.url = viper.GetString("ofd.url")
}

//RegisterReceipt sends receipt info to ekam
func (e *ekam) RegisterReceipt(cid, t string, sum float32) error {

	var rcptLines ReceiptLines
	var rcpt ReceiptRequest

	ui := billing.Billing.GetUserInfo(cid)
	if ui != nil {
		rcpt.Email = ui.Email
  		rcpt.Phone_number = ui.PhoneNumber
	}
	rcptLines.Price = sum
      	rcptLines.Quantity = 1
      	rcptLines.Title = "Услуги"
      	rcptLines.Total_price = sum
      	//rcptLines.Vat_rate    

  	//rcpt.Order_id = pid
  	//rcpt.Order_number    string
  	rcpt.Type = t
  	rcpt.Should_print = false
  	rcpt.Cash_amount = 0
  	rcpt.Electron_amount = sum
  	//rcpt.Cashier_name    string
  	rcpt.Draft = true
  	rcpt.Lines[0] = rcptLines

	jsonValue, err := json.Marshal(rcpt)
	if err != nil {
		log.Error("Ekam: " + err.Error())
		return err
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
                        return errors.New(resp.Status)
		default:
                        return errors.New(resp.Status)
                }
        } else {
		log.Error("Ekam: o response from ekam")
                return errors.New("No response from ekam")
        }

        return nil
}


