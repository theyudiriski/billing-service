package billing

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
)

var (
	ten int = 10

	CurrencyIDR = "IDR"
)

const DefaultDecimalPrecision = 2

type Amount struct {
	Val              int    `json:"value"`
	DecimalPrecision int    `json:"decimal_precision"`
	Currency         string `json:"currency"`
}

func NewAmount(val float64) Amount {
	formmatter := fmt.Sprintf("%%.%df", DefaultDecimalPrecision)
	v1 := strings.Replace(fmt.Sprintf(formmatter, val), ".", "", 1)
	v2, _ := strconv.Atoi(v1)
	return Amount{
		Val:              v2,
		DecimalPrecision: DefaultDecimalPrecision,
		Currency:         CurrencyIDR,
	}
}

func (a Amount) Value() (driver.Value, error) {
	return json.Marshal(a)
}

func (a *Amount) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("Amount.Scan error: type assertion to []byte failed")
	}

	return json.Unmarshal(b, &a)
}

func (a Amount) String() string {
	return strconv.FormatFloat(a.ToFloat64(), 'f', -1, 64)
}

func (a Amount) EqualTo(b Amount) (bool, error) {
	if a.Currency != b.Currency {
		return false, NewError(
			ErrValidationError.Error(),
			"EqualTo error: amounts needs to have same currency",
			http.StatusUnprocessableEntity,
		)
	}

	return a.ToFloat64() == b.ToFloat64(), nil
}

func (a Amount) ToFloat64() float64 {
	return float64(a.Val) / math.Pow(float64(ten), float64(a.DecimalPrecision))
}
