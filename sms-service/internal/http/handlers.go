package http

import (
	"fmt"
	"net/http"

	"notif/config"
	"notif/internal/app"
	"notif/internal/models"
	"notif/internal/repo"

	"github.com/labstack/echo/v4"
)

type ReqBody struct {
	Message     string `json:"message"`
	PhoneNumber string `json:"phone_number"`
	Org         string `json:"org"`
	IsExpress   bool   `json:"is_express"`
}

func SendSMS(c echo.Context) error {

	var req ReqBody
	cnf := config.C

	if err := c.Bind(&req); err != nil {
		panic(err)
	}

	amount := cnf.Price.Normal
	if req.IsExpress {
		amount = cnf.Price.Express
	}

	breaker := getOrgBreaker(req.Org)

	_, err := breaker.Execute(func() (any, error) {
		_, err := app.A.CreditSrv.Debit(c.Request().Context(), req.Org, amount)
		return nil, err
	})

	if err != nil {
		fmt.Printf("trying to reserve credit: %v\n", err)
		return err
	}

	log := models.SMSLog{
		PhoneNumber: req.PhoneNumber,
		Org:         req.Org,
		IsExpress:   req.IsExpress,
	}

	err = app.A.Publisher.Publish(c.Request().Context(), log)
	if err != nil {
		fmt.Printf("Could not publish on rabbit, err: %v\n", err)
	}

	return echo.NewHTTPError(200, "Ok")
}

// GetOrgCount handles GET /orgs/:org/count and returns the number of records for the org
func GetOrgCount(c echo.Context) error {
	org := c.Param("org")
	if org == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "org is required")
	}

	storage := repo.NewSMSStorage(*app.A.Db)
	count, err := storage.CountByOrg(c.Request().Context(), org)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("failed to get count: %v", err))
	}

	return c.JSON(http.StatusOK, map[string]any{
		"org":   org,
		"count": count,
	})
}

// server.go
