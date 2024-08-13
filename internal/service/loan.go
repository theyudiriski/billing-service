package billing

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type LoanService interface {
	CreateLoan(
		ctx context.Context,
		borrowerID string,
		principalAmount Amount,
		interestRate float64,
		paymentFrequency LoanFrequency,
		totalPayments int,
	) (*Loan, error)
	GetOutstanding(ctx context.Context, loanID string) (*OutstandingLoan, error)
	IsDelinquent(ctx context.Context, loanID string) (bool, error)
	GetTotalPending(ctx context.Context, loanID string) (*PendingLoan, error)
	PayLoan(ctx context.Context, loanID string, payAmount Amount) error
}

type LoanStore interface {
	CreateLoan(ctx context.Context, loan *Loan) error
	GetLoanByID(ctx context.Context, loanID string) (*Loan, error)
	GetOutstanding(ctx context.Context, loanID string) (*Amount, error)
	IsDelinquent(ctx context.Context, userID string) (bool, error)
	GetTotalPending(ctx context.Context, loanID string) (*Amount, error)
	MarkPendingAsPaid(ctx context.Context, loanID string) error
}

func NewLoanService(logger Logger, loanStore LoanStore) LoanService {
	return &loanService{
		logger:    logger,
		loanStore: loanStore,
	}
}

type loanService struct {
	logger    Logger
	loanStore LoanStore
}

type Loan struct {
	ID               string
	BorrowerID       string
	PrincipalAmount  Amount
	InterestRate     float64
	StartedAt        time.Time
	EndedAt          time.Time
	PaymentFrequency LoanFrequency
	TotalPayments    int

	// for schedules
	LoanTermDays int
	TermAmount   Amount
}

type LoanSchedule struct {
	ID        string
	AmountDue Amount
}

type (
	LoanFrequency string
)

var (
	LoanFrequencyWeekly LoanFrequency = "weekly"

	LoanFrequencies = []LoanFrequency{
		LoanFrequencyWeekly,
	}
)

func (l *LoanFrequency) UnmarshalText(text []byte) error {
	for _, frequency := range LoanFrequencies {
		if strings.EqualFold(string(frequency), string(text)) {
			*l = frequency
			return nil
		}
	}
	return NewError(
		ErrValidationError.Error(),
		fmt.Sprintf("LoanFrequency should be one of %v", LoanFrequencies),
		http.StatusBadRequest,
	)
}

type OutstandingLoan struct {
	ID     string `json:"id"`
	Amount string `json:"outstanding_amount"`
}

type PendingLoan struct {
	ID     string `json:"id"`
	Amount string `json:"pending_amount"`
}

// ! For now we always assume borrower_id is valid, despite it is random UUID

func (s *loanService) CreateLoan(
	ctx context.Context,
	borrowerID string,
	principalAmount Amount,
	interestRate float64,
	paymentFrequency LoanFrequency,
	totalPayments int,
) (*Loan, error) {
	principalFloat := principalAmount.ToFloat64()

	// calculate total amount & term amount
	totalAmount := principalFloat * (1 + interestRate)
	payment := totalAmount / float64(totalPayments)
	termAmount := NewAmount(payment)

	// calculate loan term in days
	loanTermDays := convertLoanTermToDays(paymentFrequency, totalPayments)

	// calculate start and end date
	start := CurrentLocalTime()
	end := start.AddDate(0, 0, loanTermDays)

	loan := &Loan{
		ID:               UUID(),
		BorrowerID:       borrowerID,
		PrincipalAmount:  principalAmount,
		InterestRate:     interestRate,
		StartedAt:        start,
		EndedAt:          end,
		PaymentFrequency: paymentFrequency,
		TotalPayments:    totalPayments,

		LoanTermDays: loanTermDays,
		TermAmount:   termAmount,
	}

	if err := s.loanStore.CreateLoan(ctx, loan); err != nil {
		s.logger.WarnContext(ctx, "failed to create loan", "error", err)
		return nil, err
	}

	return loan, nil
}

func convertLoanTermToDays(
	paymentFrequency LoanFrequency,
	totalPayments int,
) int {
	// currently only support weekly
	switch paymentFrequency {
	case LoanFrequencyWeekly:
		return totalPayments * 7
	default:
		return 0
	}
}

func (s *loanService) GetOutstanding(
	ctx context.Context,
	loanID string,
) (*OutstandingLoan, error) {
	_, err := s.loanStore.GetLoanByID(ctx, loanID)
	if err != nil {
		s.logger.WarnContext(ctx, "failed to get loan", "error", err)
		return nil, err
	}

	outstandingAmount, err := s.loanStore.GetOutstanding(ctx, loanID)
	if err != nil {
		s.logger.WarnContext(ctx, "failed to get outstanding loan", "error", err)
		return nil, err
	}

	return &OutstandingLoan{
		ID:     loanID,
		Amount: outstandingAmount.String(),
	}, nil
}

func (s *loanService) IsDelinquent(
	ctx context.Context,
	loanID string,
) (bool, error) {
	_, err := s.loanStore.GetLoanByID(ctx, loanID)
	if err != nil {
		s.logger.WarnContext(ctx, "failed to get loan", "error", err)
		return false, err
	}

	return s.loanStore.IsDelinquent(ctx, loanID)
}

func (s *loanService) GetTotalPending(
	ctx context.Context,
	loanID string,
) (*PendingLoan, error) {
	_, err := s.loanStore.GetLoanByID(ctx, loanID)
	if err != nil {
		s.logger.WarnContext(ctx, "failed to get loan", "error", err)
		return nil, err
	}

	pendingAmount, err := s.loanStore.GetTotalPending(ctx, loanID)
	if err != nil {
		s.logger.WarnContext(ctx, "failed to get pending loan", "error", err)
		return nil, err
	}

	return &PendingLoan{
		ID:     loanID,
		Amount: pendingAmount.String(),
	}, nil
}

func (s *loanService) PayLoan(ctx context.Context, loanID string, payAmount Amount) error {
	loan, err := s.loanStore.GetLoanByID(ctx, loanID)
	if err != nil {
		s.logger.WarnContext(ctx, "failed to get loan", "error", err)
		return err
	}

	pendingAmount, err := s.loanStore.GetTotalPending(ctx, loan.ID)
	if err != nil {
		s.logger.WarnContext(ctx, "failed to get pending loan", "error", err)
		return err
	}

	isEqual, err := payAmount.EqualTo(*pendingAmount)
	if err != nil {
		s.logger.WarnContext(ctx, "failed to compare amount", "error", err)
		return err
	}

	if !isEqual {
		s.logger.WarnContext(ctx, "payment amount mismatch", "payAmount", payAmount, "pendingAmount", pendingAmount)
		return ErrPaymentAmountMismatch
	}

	if err := s.loanStore.MarkPendingAsPaid(ctx, loan.ID); err != nil {
		s.logger.WarnContext(ctx, "failed to mark loan as paid", "error", err)
		return err
	}

	// TODO #1: implement create payment record to save payment history
	// related with payment amount, payment date, 3rd party payment id, etc.

	// TODO #2: change loan status to paid if all pending payment has been paid.

	return nil
}
