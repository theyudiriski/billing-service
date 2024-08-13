package billing

import "errors"

var (
	ErrServerError               error = errors.New("SERVER_ERROR")
	ErrValidationError           error = errors.New("VALIDATION_ERROR")
	ErrUnprocessableContentError error = errors.New("UNPROCESSABLE_CONTENT_ERROR")

	ErrInvalidUUID error = errors.New("INVALID_UUID")

	ErrLoanNotFound          error = errors.New("LOAN_NOT_FOUND")
	ErrPaymentAmountMismatch error = errors.New("PAYMENT_AMOUNT_MISMATCH")
)
