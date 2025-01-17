package controllers

import (
	"context"
	"net/http"

	"github.com/getAlby/lndhub.go/lib"
	"github.com/getAlby/lndhub.go/lib/service"
	"github.com/labstack/echo/v4"
)

// GetTXSController : GetTXSController struct
type GetTXSController struct {
	svc *service.LndhubService
}

func NewGetTXSController(svc *service.LndhubService) *GetTXSController {
	return &GetTXSController{svc: svc}
}

// GetTXS : Get TXS Controller
func (controller *GetTXSController) GetTXS(c echo.Context) error {
	userId := c.Get("UserID").(int64)

	invoices, err := controller.svc.InvoicesFor(context.TODO(), userId, "outgoing")
	if err != nil {
		return err
	}

	response := make([]echo.Map, len(invoices))
	for i, invoice := range invoices {
		rhash, _ := lib.ToJavaScriptBuffer(invoice.RHash)
		response[i] = echo.Map{
			"r_hash":           rhash,
			"payment_hash":     rhash,
			"payment_preimage": invoice.Preimage,
			"value":            invoice.Amount,
			"type":             "paid_invoice",
			"fee":              0, //TODO charge fees
			"timestamp":        invoice.CreatedAt.Unix(),
			"memo":             invoice.Memo,
		}
	}
	return c.JSON(http.StatusOK, &response)
}

func (controller *GetTXSController) GetUserInvoices(c echo.Context) error {
	userId := c.Get("UserID").(int64)

	invoices, err := controller.svc.InvoicesFor(context.TODO(), userId, "incoming")
	if err != nil {
		return err
	}

	response := make([]echo.Map, len(invoices))
	for i, invoice := range invoices {
		rhash, _ := lib.ToJavaScriptBuffer(invoice.RHash)
		response[i] = echo.Map{
			"r_hash":          rhash,
			"payment_request": invoice.PaymentRequest,
			"pay_req":         invoice.PaymentRequest,
			"description":     invoice.Memo,
			"payment_hash":    invoice.RHash,
			"ispaid":          invoice.State == "settled",
			"amt":             invoice.Amount,
			"expire_time":     3600 * 24,
			"timestamp":       invoice.CreatedAt.Unix(),
			"type":            "user_invoice",
		}
	}
	return c.JSON(http.StatusOK, &response)
}
