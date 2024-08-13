package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	billing "github.com/theyudiriski/billing-service/internal/service"
)

const (
	delinquencyThreshold = 2
)

func NewLoanStore(db *Client) billing.LoanStore {
	return &loanStore{db}
}

type loanStore struct {
	db *Client
}

func (s *loanStore) CreateLoan(
	ctx context.Context,
	loan *billing.Loan,
) error {
	tx, err := s.db.Leader.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
INSERT INTO loans(
	id,
	borrower_id,
	principal_amount,
	interest_rate,
	started_at,
	ended_at,
	payment_frequency,
	total_payments
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		loan.ID,
		loan.BorrowerID,
		loan.PrincipalAmount,
		loan.InterestRate,
		loan.StartedAt,
		loan.EndedAt,
		loan.PaymentFrequency,
		loan.TotalPayments,
	)
	if err != nil {
		return err
	}

	stmt, err := tx.PrepareContext(ctx, `
	INSERT INTO loan_schedules(
		id,
		loan_id,
		seq,
		due_date,
		amount_due
	) VALUES ($1, $2, $3, $4, $5)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i := 1; i <= loan.TotalPayments; i++ {
		scheduleID := billing.UUID()

		// calculate due date for each period
		days := loan.LoanTermDays * i / loan.TotalPayments
		dueDate := loan.StartedAt.AddDate(0, 0, days)

		_, err := stmt.Exec(
			scheduleID,
			loan.ID,
			i,
			dueDate,
			loan.TermAmount,
		)
		if err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

// GetOutstanding returns the total amount of outstanding payments for a loan.
// It calculates the total amount due from unpaid loan schedules.
func (s *loanStore) GetOutstanding(
	ctx context.Context,
	loanID string,
) (*billing.Amount, error) {
	var totalDue float64
	tx, err := s.db.Leader.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	// fetch total amount due from unpaid loans
	row := tx.QueryRowContext(ctx, `
SELECT 
	COALESCE(
		SUM(
			CAST(amount_due->>'value' AS BIGINT) / 
			POWER(10, CAST(amount_due->>'decimal_precision' AS INTEGER))
		), 
		0
	) AS total_pending
FROM 
	loan_schedules
WHERE 
	loan_id = $1 
	AND status = 'unpaid'`,
		loanID,
	)

	if err := row.Scan(&totalDue); err != nil {
		return nil, err
	}

	amount := billing.NewAmount(totalDue)
	return &amount, nil
}

func (s *loanStore) IsDelinquent(ctx context.Context, loanID string) (bool, error) {
	tx, err := s.db.Leader.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}

	rows, err := tx.QueryContext(ctx, `
SELECT
	due_date
FROM
	loan_schedules
WHERE
	loan_id = $1 AND
	status = 'unpaid'
ORDER BY
	due_date
    `, loanID)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	var dueDates []time.Time
	for rows.Next() {
		var dueDate time.Time
		if err := rows.Scan(&dueDate); err != nil {
			return false, err
		}
		dueDates = append(dueDates, dueDate)
	}
	if err := rows.Err(); err != nil {
		return false, err
	}

	missedWeeks := calculateMissedWeeks(dueDates)

	return missedWeeks > delinquencyThreshold, nil
}

func calculateMissedWeeks(dueDates []time.Time) int {
	// current time
	now := billing.CurrentLocalTime()

	missedWeeks := 0
	for i := range dueDates {
		// TODO: checker should only care about the date, not the time,
		// but both of them must be in the same timezone.
		if now.After(dueDates[i]) {
			missedWeeks++
		} else {
			break
		}
	}

	return missedWeeks
}

// GetTotalPending returns the total amount of pending payments for a loan that are past due date.
func (s *loanStore) GetTotalPending(ctx context.Context, loanID string) (*billing.Amount, error) {
	var totalPending float64
	tx, err := s.db.Leader.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	row := tx.QueryRowContext(ctx, `
SELECT 
	COALESCE(
		SUM(
			CAST(amount_due->>'value' AS BIGINT) / 
			POWER(10, CAST(amount_due->>'decimal_precision' AS INTEGER))
		), 
		0
	) AS total_pending
FROM 
	loan_schedules
WHERE 
	loan_id = $1 
	AND status = 'unpaid'
	AND due_date < NOW()`,
		loanID,
	)

	err = row.Scan(&totalPending)
	if err != nil {
		return nil, err
	}

	amount := billing.NewAmount(totalPending)
	return &amount, nil
}

func (s *loanStore) MarkPendingAsPaid(ctx context.Context, loanID string) error {
	tx, err := s.db.Leader.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
UPDATE 
	loan_schedules
SET 
	status = 'paid'
WHERE 
	loan_id = $1 
	AND status = 'unpaid'
	AND due_date < NOW()`,
		loanID,
	)
	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (s *loanStore) GetLoanByID(ctx context.Context, loanID string) (*billing.Loan, error) {
	tx, err := s.db.Leader.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	l := &billing.Loan{}
	row := tx.QueryRowContext(ctx, `
SELECT
	id,
	borrower_id,
	principal_amount,
	interest_rate,
	started_at,
	ended_at,
	payment_frequency,
	total_payments
FROM
	loans
WHERE
	id = $1`,
		loanID)
	if err = row.Scan(
		&l.ID,
		&l.BorrowerID,
		&l.PrincipalAmount,
		&l.InterestRate,
		&l.StartedAt,
		&l.EndedAt,
		&l.PaymentFrequency,
		&l.TotalPayments,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, billing.ErrLoanNotFound
		}
		return nil, err
	}

	return l, nil
}
