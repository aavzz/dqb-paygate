package ofd

type ReceiptLines struct {
      Price      float32 `json:"price"`
      Quantity   int     `json:"quantity"`
      Title      string  `json:"title"`
      TotalPrice float32 `json:"total_price"`
      VatRate    *int64  `json:"vat_rate"`
}

type ReceiptRequest struct {
  OrderId        string         `json:"order_id"`
  OrderNumber    string         `json:"order_number"`
  Type           string         `json:"type"`
  Email          string         `json:"email,omitempty"`
  PhoneNumber    string         `json:"phone_number,omitempty"`
  ShouldPrint    bool           `json:"should_print"`
  CashAmount     float32        `json:"cash_amount"`
  ElectronAmount float32        `json:"electron_amount"`
  CashierName    string         `json:"cashier_name,omitempty"`
  Draft          bool           `json:"draft"`
  Lines          []ReceiptLines `json:"lines"`
}

type ResponseOkLines struct {
      Id          uint64 `json:"id"`
      Title       string `json:"title"`
      Quantity    string `json:"quantity"`
      TotalPrice  string `json:"total_price"`
      Price       string `json:"price"`
      VatRate     uint8  `json:"vat_rate"`
      VatAmount   string `json:"vat_amount"`
}

type ResponseOkFiscalData struct {
    ReceiptNumber      uint64 `json:"receipt_number"`
    ModelNumber        string `json:"model_number"`
    FactoryKktNumber   string `json:"factory_kkt_number"`
    FactoryFnNumber    string `json:"factory_fn_number"`
    RegistrationNumber string `json:"registration_number"`
    FnExpiredPeriod    uint   `json:"fn_expired_period"`
    FdNumber           uint   `json:"fd_number"`
    Fpd                uint   `json:"fpd"`
    TaxSystem          string `json:"tax_system"`
    OrganizationName   string `json:"organization_name"`
    OrganizationInn    string `json:"organization_inn"`
    Address            string `json:"address"`
    RetailShiftNumber  string `json:"retail_shift_number"`
    OfdName            string `json:"ofd_name"`
    PrintedAt          string `json:"printed_at"`
    RegistrationDate   string `json:"registration_date"`
    FnExpiredAt        string `json:"fn_expired_at"`
}

type ResponseOk struct {
  Id                 uint64               `json:"id"`
  Uuid               string               `json:"uuid"`
  Type               string               `json:"type"`
  Status             string               `json:"status"`
  KktReceiptId       uint                 `json:"kkt_receipt_id"`
  Amount             string               `json:"amount"`
  CashAmount         string               `json:"cash_amount"`
  ElectronAmount     string               `json:"electron_amount"`
  Lines              []ResponseOkLines    `json:"lines"`
  CashierName        string               `json:"cashier_name"`
  CashierRole        string               `json:"cashier_role"`
  CashierInn         string               `json:"cashier_inn"`
  TransactionAddress string               `json:"transaction_address"`
  Email              string               `json:"email"`
  PhoneNumber        string               `json:"phone_number"`
  ShouldPrint        bool                 `json:"should_print"`
  OrderId            string               `json:"order_id"`
  OrderNumber        string               `json:"order_number"`
  CreatedAt          string               `json:"created_at"`
  UpdatedAt          string               `json:"updated_at"`
  KktReceiptExists   bool                 `json:"kkt_receipt_exists"`
  Draft              bool                 `json:"draft"`
  Copy               bool                 `json:"copy"`
  FiscalData         ResponseOkFiscalData `json:"fiscal_data"`
  ReceiptUrl         string               `json:"receipt_url"`
  OnlineCashierUrl   string               `json:"online_cachier_url"`
  Error              string               `json:"error"`
}

type ResponseError struct {
	Error string `json:"error"`
}

