package integration_tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/getAlby/lndhub.go/controllers"
	"github.com/getAlby/lndhub.go/lib"
	"github.com/getAlby/lndhub.go/lib/responses"
	"github.com/getAlby/lndhub.go/lib/service"
	"github.com/getAlby/lndhub.go/lib/tokens"
	"github.com/getAlby/lndhub.go/lnd"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type PaymentTestSuite struct {
	suite.Suite
	echo          *echo.Echo
	fundingClient lnrpc.LightningClient
	service       *service.LndhubService
	aliceLogin    controllers.CreateUserResponseBody
	aliceToken    string
	bobLogin      controllers.CreateUserResponseBody
	bobToken      string
}

func (suite *PaymentTestSuite) SetupSuite() {
	lndClient, err := lnd.NewLNDclient(lnd.LNDoptions{
		Address:     lnd2RegtestAddress,
		MacaroonHex: lnd2RegtestMacaroonHex,
	})
	if err != nil {
		log.Fatalf("Error setting up funding client: %v", err)
	}
	suite.fundingClient = lndClient

	svc, users, userTokens, err := LndHubTestServiceInit(2)
	if err != nil {
		log.Fatalf("Error initializing test service: %v", err)
	}
	// Subscribe to LND invoice updates in the background
	go svc.InvoiceUpdateSubscription(context.Background())
	suite.service = svc
	e := echo.New()

	e.HTTPErrorHandler = responses.HTTPErrorHandler
	e.Validator = &lib.CustomValidator{Validator: validator.New()}
	suite.echo = e
	assert.Equal(suite.T(), 2, len(users))
	assert.Equal(suite.T(), 2, len(userTokens))
	suite.aliceLogin = users[0]
	suite.aliceToken = userTokens[0]
	suite.bobLogin = users[1]
	suite.bobToken = userTokens[1]
	suite.echo.Use(tokens.Middleware([]byte(suite.service.Config.JWTSecret)))
	suite.echo.GET("/balance", controllers.NewBalanceController(suite.service).Balance)
	suite.echo.POST("/addinvoice", controllers.NewAddInvoiceController(suite.service).AddInvoice)
	suite.echo.POST("/payinvoice", controllers.NewPayInvoiceController(suite.service).PayInvoice)
}

func (suite *PaymentTestSuite) TearDownSuite() {

}

func (suite *PaymentTestSuite) createAddInvoiceReq(amt int, memo, token string) *controllers.AddInvoiceResponseBody {
	rec := httptest.NewRecorder()
	fundingSatAmt := 1000
	var buf bytes.Buffer
	assert.NoError(suite.T(), json.NewEncoder(&buf).Encode(&controllers.AddInvoiceRequestBody{
		Amount: fundingSatAmt,
		Memo:   "integration test IncomingPaymentTestSuite",
	}))
	req := httptest.NewRequest(http.MethodPost, "/addinvoice", &buf)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
	suite.echo.ServeHTTP(rec, req)
	invoiceResponse := &controllers.AddInvoiceResponseBody{}
	assert.Equal(suite.T(), http.StatusOK, rec.Code)
	assert.NoError(suite.T(), json.NewDecoder(rec.Body).Decode(invoiceResponse))
	return invoiceResponse
}

func (suite *PaymentTestSuite) TestInternalPayment() {
	aliceFundingSats := 1000
	bobSatRequested := 500
	//fund alice account
	invoiceResponse := suite.createAddInvoiceReq(aliceFundingSats, "integration test internal payment alice", suite.aliceToken)
	sendPaymentRequest := lnrpc.SendRequest{
		PaymentRequest: invoiceResponse.PayReq,
		FeeLimit:       nil,
	}
	_, err := suite.fundingClient.SendPaymentSync(context.TODO(), &sendPaymentRequest)
	assert.NoError(suite.T(), err)

	//wait a bit for the callback event to hit
	time.Sleep(100 * time.Millisecond)

	//create invoice for bob
	bobInvoice := suite.createAddInvoiceReq(bobSatRequested, "integration test internal payment bob", suite.bobToken)
	//pay bob from alice
	rec := httptest.NewRecorder()
	var buf bytes.Buffer
	assert.NoError(suite.T(), json.NewEncoder(&buf).Encode(&controllers.PayInvoiceRequestBody{
		Invoice: bobInvoice.PayReq,
	}))
	req := httptest.NewRequest(http.MethodPost, "/payinvoice", &buf)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", suite.aliceToken))
	suite.echo.ServeHTTP(rec, req)
	payResponse := &controllers.PayInvoiceResponseBody{}
	assert.Equal(suite.T(), http.StatusOK, rec.Code)
	assert.NoError(suite.T(), json.NewDecoder(rec.Body).Decode(payResponse))
	assert.NotEmpty(suite.T(), payResponse.PaymentPreimage)

}

func TestInternalPaymentTestSuite(t *testing.T) {
	suite.Run(t, new(PaymentTestSuite))
}
