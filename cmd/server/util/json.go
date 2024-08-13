package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	billing "github.com/theyudiriski/billing-service/internal/service"
)

var errorMapping = map[error]error{
	billing.ErrInvalidUUID: billing.NewError(
		billing.ErrInvalidUUID.Error(),
		"UUID provided is invalid",
		http.StatusBadRequest,
	),

	billing.ErrLoanNotFound: billing.NewError(
		billing.ErrLoanNotFound.Error(),
		"Loan not found",
		http.StatusBadRequest,
	),

	billing.ErrPaymentAmountMismatch: billing.NewError(
		billing.ErrPaymentAmountMismatch.Error(),
		"Payment amount is not equal to pending amount",
		http.StatusBadRequest,
	),
}

func MarshalJSONResponse(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		fmt.Printf("MarshalJSONResponse [status] %v [error] %v\n", statusCode, err)
	}
}

func MarshalJSONSuccess(w http.ResponseWriter, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := map[string]string{
		"status": "success",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		fmt.Printf("MarshalJSONSuccess [status] %v [error] %v\n", statusCode, err)
	}
}

func MarshalJSONError(w http.ResponseWriter, err error) {
	logger := billing.NewLogger()
	logger.Warn(fmt.Sprintf("MarshalJSONError: %v", err.Error()))

	err = handleError(err)

	var cErr billing.CustomError
	if errors.As(err, &cErr) {
		MarshalJSONResponse(w, cErr.HTTPStatusCode(), map[string]string{
			"error_code": cErr.Code(),
			"message":    err.Error(),
		})
		return
	}

	MarshalJSONResponse(w, http.StatusInternalServerError, map[string]string{
		"error_code": billing.ErrServerError.Error(),
		"message":    "Something went wrong, please try again later.",
	})
}

func handleError(err error) error {
	rootError := unwrapError(err)
	mappedError, ok := errorMapping[rootError]
	if !ok {
		return err
	}
	return mappedError
}

func unwrapError(err error) error {
	rootErr := errors.Unwrap(err)
	if rootErr == nil {
		return err
	}
	return unwrapError(rootErr)
}
