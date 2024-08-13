package billing_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	mock_billing "github.com/theyudiriski/billing-service/internal/service/mock"

	billing "github.com/theyudiriski/billing-service/internal/service"

	. "github.com/smartystreets/goconvey/convey"
)

var (
	mockLoanStore *mock_billing.MockLoanStore

	loanService billing.LoanService
	errMock     error = errors.New("mock error")
)

func provideLoanTest(t *testing.T) func() {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLoanStore = mock_billing.NewMockLoanStore(ctrl)

	loanService = billing.NewLoanService(
		billing.NewLogger(),
		mockLoanStore,
	)

	return func() {}
}

func TestCreateLoan(t *testing.T) {
	finish := provideLoanTest(t)
	defer finish()

	Convey("CreateLoan", t, FailureHalts, func() {
		type (
			args struct {
				ctx              context.Context
				borrowerID       string
				principalAmount  billing.Amount
				interestRate     float64
				paymentFrequency billing.LoanFrequency
				totalPayments    int
			}
		)

		var (
			ctx              = context.Background()
			borrowerID       = "borrower-id"
			principalAmount  = billing.NewAmount(5_000_000)
			interestRate     = 0.1
			paymentFrequency = billing.LoanFrequencyWeekly
			totalPayments    = 50
		)

		testCases := []struct {
			testID   int
			testDesc string
			testType string
			args     args
			mock     func()
		}{
			{
				testID:   1,
				testDesc: "success create loan",
				testType: "P",
				args: args{
					ctx:              ctx,
					borrowerID:       borrowerID,
					principalAmount:  principalAmount,
					interestRate:     interestRate,
					paymentFrequency: paymentFrequency,
					totalPayments:    totalPayments,
				},
				mock: func() {
					mockLoanStore.EXPECT().CreateLoan(ctx, gomock.Any()).
						Do(func(ctx context.Context, loan *billing.Loan) {
							So(loan.BorrowerID, ShouldEqual, borrowerID)
							So(loan.PrincipalAmount, ShouldEqual, principalAmount)
							So(loan.InterestRate, ShouldEqual, interestRate)
							So(loan.PaymentFrequency, ShouldEqual, paymentFrequency)
							So(loan.TotalPayments, ShouldEqual, totalPayments)

							So(loan.LoanTermDays, ShouldEqual, 7*totalPayments)
							So(loan.TermAmount, ShouldEqual, billing.NewAmount(110_000))
						}).Return(nil)
				},
			},
			{
				testID:   2,
				testDesc: "failed create loan",
				testType: "N",
				args: args{
					ctx:              ctx,
					borrowerID:       borrowerID,
					principalAmount:  principalAmount,
					interestRate:     interestRate,
					paymentFrequency: paymentFrequency,
					totalPayments:    totalPayments,
				},
				mock: func() {
					mockLoanStore.EXPECT().CreateLoan(ctx, gomock.Any()).Return(errMock)
				},
			},
		}

		for _, tc := range testCases {
			t.Logf("%d - [%s] : %s", tc.testID, tc.testType, tc.testDesc)
			tc.mock()

			_, err := loanService.CreateLoan(
				tc.args.ctx,
				tc.args.borrowerID,
				tc.args.principalAmount,
				tc.args.interestRate,
				tc.args.paymentFrequency,
				tc.args.totalPayments,
			)

			if tc.testType == "P" {
				So(err, ShouldBeNil)
			} else {
				So(err, ShouldNotBeNil)
			}
		}
	})
}

func TestPayLoan(t *testing.T) {
	finish := provideLoanTest(t)
	defer finish()

	Convey("PayLoan", t, FailureHalts, func() {
		type (
			args struct {
				ctx       context.Context
				loanID    string
				payAmount billing.Amount
			}
		)

		var (
			ctx       = context.Background()
			loanID    = "loan-id"
			payAmount = billing.NewAmount(100)

			oneHundredAmount  = billing.NewAmount(100)
			oneThousandAmount = billing.NewAmount(1_000)

			loan = billing.Loan{
				ID: loanID,
			}
		)

		testCases := []struct {
			testID      int
			testDesc    string
			testType    string
			args        args
			expectedErr error
			mock        func()
		}{
			{
				testID:   1,
				testDesc: "success pay loan",
				testType: "P",
				args: args{
					ctx:       ctx,
					loanID:    loanID,
					payAmount: payAmount,
				},
				mock: func() {
					mockLoanStore.EXPECT().GetLoanByID(ctx, loanID).Return(&loan, nil)
					mockLoanStore.EXPECT().GetTotalPending(ctx, loanID).Return(&oneHundredAmount, nil)
					mockLoanStore.EXPECT().MarkPendingAsPaid(ctx, loanID).Return(nil)
				},
			},
			{
				testID:   2,
				testDesc: "failed: loan not found",
				testType: "N",
				args: args{
					ctx:       ctx,
					loanID:    loanID,
					payAmount: payAmount,
				},
				mock: func() {
					mockLoanStore.EXPECT().GetLoanByID(ctx, loanID).Return(nil, billing.ErrLoanNotFound)
				},
				expectedErr: billing.ErrLoanNotFound,
			},
			{
				testID:   3,
				testDesc: "failed: get total pending",
				testType: "N",
				args: args{
					ctx:       ctx,
					loanID:    loanID,
					payAmount: payAmount,
				},
				mock: func() {
					mockLoanStore.EXPECT().GetLoanByID(ctx, loanID).Return(&loan, nil)
					mockLoanStore.EXPECT().GetTotalPending(ctx, loanID).Return(nil, errMock)
				},
				expectedErr: errMock,
			},
			{
				testID:   4,
				testDesc: "failed: payment amount mismatch",
				testType: "N",
				args: args{
					ctx:       ctx,
					loanID:    loanID,
					payAmount: payAmount,
				},
				mock: func() {
					mockLoanStore.EXPECT().GetLoanByID(ctx, loanID).Return(&loan, nil)
					mockLoanStore.EXPECT().GetTotalPending(ctx, loanID).Return(&oneThousandAmount, nil)
				},
				expectedErr: billing.ErrPaymentAmountMismatch,
			},
			{
				testID:   5,
				testDesc: "failed: mark pending as paid",
				testType: "N",
				args: args{
					ctx:       ctx,
					loanID:    loanID,
					payAmount: payAmount,
				},
				mock: func() {
					mockLoanStore.EXPECT().GetLoanByID(ctx, loanID).Return(&loan, nil)
					mockLoanStore.EXPECT().GetTotalPending(ctx, loanID).Return(&oneHundredAmount, nil)
					mockLoanStore.EXPECT().MarkPendingAsPaid(ctx, loanID).Return(errMock)
				},
				expectedErr: errMock,
			},
		}

		for _, tc := range testCases {
			t.Logf("%d - [%s] : %s", tc.testID, tc.testType, tc.testDesc)
			tc.mock()

			err := loanService.PayLoan(
				tc.args.ctx,
				tc.args.loanID,
				tc.args.payAmount,
			)

			if tc.testType == "P" {
				So(err, ShouldBeNil)
			} else {
				So(err, ShouldNotBeNil)
				So(err, ShouldEqual, tc.expectedErr)
			}
		}
	})
}
