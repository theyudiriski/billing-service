package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/theyudiriski/billing-service/cmd/server/util"
	billing "github.com/theyudiriski/billing-service/internal/service"
)

// CreateLoan
type CreateLoanRequest struct {
	BorrowerID       string
	PrincipalAmount  billing.Amount
	InterestRate     float64
	PaymentFrequency billing.LoanFrequency
	TotalPayments    int
}

func (r *CreateLoanRequest) UnmarshalJSON(b []byte) error {
	temp := struct {
		BorrowerID       *string                `json:"borrower_id"`
		PrincipalAmount  *float64               `json:"principal_amount"`
		InterestRate     *float64               `json:"interest_rate"`
		PaymentFrequency *billing.LoanFrequency `json:"payment_frequency"`
		TotalPayments    *int                   `json:"total_payments"`
	}{}

	if err := json.Unmarshal(b, &temp); err != nil {
		return billing.NewError(
			billing.ErrValidationError.Error(),
			err.Error(),
			http.StatusBadRequest,
		)
	}

	if temp.BorrowerID == nil || *temp.BorrowerID == "" {
		return billing.NewError(
			billing.ErrValidationError.Error(),
			"borrower_id is required",
			http.StatusBadRequest,
		)
	}

	if temp.PrincipalAmount == nil || *temp.PrincipalAmount <= 0 {
		return billing.NewError(
			billing.ErrValidationError.Error(),
			"principal_amount is required and must be greater than 0",
			http.StatusBadRequest,
		)
	}

	if temp.InterestRate == nil || *temp.InterestRate <= 0 {
		return billing.NewError(
			billing.ErrValidationError.Error(),
			"interest_rate is required and must be greater than 0",
			http.StatusBadRequest,
		)
	}

	if temp.PaymentFrequency == nil {
		return billing.NewError(
			billing.ErrValidationError.Error(),
			"payment_frequency is required",
			http.StatusBadRequest,
		)
	}

	if temp.TotalPayments == nil || *temp.TotalPayments <= 0 {
		return billing.NewError(
			billing.ErrValidationError.Error(),
			"total_payments is required and must be greater than 0",
			http.StatusBadRequest,
		)
	}

	*r = CreateLoanRequest{
		BorrowerID:       *temp.BorrowerID,
		PrincipalAmount:  billing.NewAmount(*temp.PrincipalAmount),
		InterestRate:     *temp.InterestRate,
		PaymentFrequency: *temp.PaymentFrequency,
		TotalPayments:    *temp.TotalPayments,
	}

	return nil
}

type LoanResponse struct {
	*billing.Loan
}

func (r LoanResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		ID               string  `json:"id"`
		BorrowerID       string  `json:"borrower_id"`
		PrincipalAmount  float64 `json:"principal_amount"`
		InterestRate     float64 `json:"interest_rate"`
		StartedAt        string  `json:"started_at"`
		EndedAt          string  `json:"ended_at"`
		PaymentFrequency string  `json:"payment_frequency"`
		TotalPayments    int     `json:"total_payments"`
	}{
		ID:               r.ID,
		BorrowerID:       r.BorrowerID,
		PrincipalAmount:  r.PrincipalAmount.ToFloat64(),
		InterestRate:     r.InterestRate,
		StartedAt:        billing.LocalTime(r.StartedAt).Format("2006-01-02"),
		EndedAt:          billing.LocalTime(r.EndedAt).Format("2006-01-02"),
		PaymentFrequency: string(r.PaymentFrequency),
		TotalPayments:    r.TotalPayments,
	})
}

func CreateLoan(
	logger billing.Logger,
	loanService billing.LoanService,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		var in CreateLoanRequest
		reqBody, err := io.ReadAll(r.Body)
		defer r.Body.Close()

		if err != nil {
			logger.WarnContext(ctx, "failed to read request body", "error", err)
			util.MarshalJSONError(w, err)
			return
		}

		if err = json.Unmarshal(reqBody, &in); err != nil {
			logger.WarnContext(ctx, "failed to unmarshal request body", "error", err)

			var syntaxError *json.SyntaxError
			if errors.As(err, &syntaxError) {
				err = billing.NewError(
					billing.ErrUnprocessableContentError.Error(),
					"Invalid json.",
					http.StatusUnprocessableEntity,
				)
			}

			util.MarshalJSONError(w, err)
			return
		}

		loan, err := loanService.CreateLoan(
			ctx,
			in.BorrowerID,
			in.PrincipalAmount,
			in.InterestRate,
			in.PaymentFrequency,
			in.TotalPayments,
		)
		if err != nil {
			logger.WarnContext(ctx, "failed to create loan", "error", err)
			util.MarshalJSONError(w, err)
			return
		}

		util.MarshalJSONResponse(w, http.StatusCreated, LoanResponse{loan})
	}
}

// GetOutstandingLoan
func GetOutstandingLoan(
	logger billing.Logger,
	loanService billing.LoanService,
	id string,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		if _, errParse := uuid.Parse(id); errParse != nil {
			util.MarshalJSONError(w, billing.ErrInvalidUUID)
			return
		}

		outstanding, err := loanService.GetOutstanding(ctx, id)
		if err != nil {
			logger.WarnContext(ctx, "failed to get outstanding loan", "error", err)
			util.MarshalJSONError(w, err)
			return
		}

		util.MarshalJSONResponse(w, http.StatusOK, outstanding)
	}
}

// IsDelinquent
type IsDelinquentResponse struct {
	LoanID       string `json:"loan_id"`
	IsDelinquent bool   `json:"is_delinquent"`
}

func IsDelinquent(
	logger billing.Logger,
	loanService billing.LoanService,
	id string,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		if _, errParse := uuid.Parse(id); errParse != nil {
			util.MarshalJSONError(w, billing.ErrInvalidUUID)
			return
		}

		isDelinquent, err := loanService.IsDelinquent(ctx, id)
		if err != nil {
			logger.WarnContext(ctx, "failed to get outstanding loan", "error", err)
			util.MarshalJSONError(w, err)
			return
		}

		response := IsDelinquentResponse{
			LoanID:       id,
			IsDelinquent: isDelinquent,
		}

		util.MarshalJSONResponse(w, http.StatusOK, response)
	}
}

// GetPendingLoan
func GetPendingLoan(
	logger billing.Logger,
	loanService billing.LoanService,
	id string,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		if _, errParse := uuid.Parse(id); errParse != nil {
			util.MarshalJSONError(w, billing.ErrInvalidUUID)
			return
		}

		pending, err := loanService.GetTotalPending(ctx, id)
		if err != nil {
			logger.WarnContext(ctx, "failed to get pending loan", "error", err)
			util.MarshalJSONError(w, err)
			return
		}

		util.MarshalJSONResponse(w, http.StatusOK, pending)
	}
}

// PayLoan
type PayLoanRequest struct {
	ID     string
	Amount billing.Amount
}

func (r *PayLoanRequest) UnmarshalJSON(b []byte) error {
	temp := struct {
		ID     *string  `json:"id"`
		Amount *float64 `json:"amount"`
	}{}

	if err := json.Unmarshal(b, &temp); err != nil {
		return billing.NewError(
			billing.ErrValidationError.Error(),
			err.Error(),
			http.StatusBadRequest,
		)
	}

	if temp.ID == nil || *temp.ID == "" {
		return billing.NewError(
			billing.ErrValidationError.Error(),
			"id is required",
			http.StatusBadRequest,
		)
	}

	if temp.Amount == nil || *temp.Amount <= 0 {
		return billing.NewError(
			billing.ErrValidationError.Error(),
			"amount is required and must be greater than 0",
			http.StatusBadRequest,
		)
	}

	*r = PayLoanRequest{
		ID:     *temp.ID,
		Amount: billing.NewAmount(*temp.Amount),
	}

	return nil
}

func PayLoan(
	logger billing.Logger,
	loanService billing.LoanService,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		var in PayLoanRequest
		reqBody, err := io.ReadAll(r.Body)
		defer r.Body.Close()

		if err != nil {
			logger.WarnContext(ctx, "failed to read request body", "error", err)
			util.MarshalJSONError(w, err)
			return
		}

		if err = json.Unmarshal(reqBody, &in); err != nil {
			logger.WarnContext(ctx, "failed to unmarshal request body", "error", err)

			var syntaxError *json.SyntaxError
			if errors.As(err, &syntaxError) {
				err = billing.NewError(
					billing.ErrUnprocessableContentError.Error(),
					"Invalid json.",
					http.StatusUnprocessableEntity,
				)
			}

			util.MarshalJSONError(w, err)
			return
		}

		err = loanService.PayLoan(
			ctx,
			in.ID,
			in.Amount,
		)
		if err != nil {
			logger.WarnContext(ctx, "failed to pay loan", "error", err)
			util.MarshalJSONError(w, err)
			return
		}

		util.MarshalJSONSuccess(w, http.StatusOK)
	}
}
