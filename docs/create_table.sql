CREATE TABLE loans (
    id                  VARCHAR(36)     NOT NULL,
    borrower_id         VARCHAR(36)     NOT NULL,
    principal_amount    JSONB           NOT NULL,
    interest_rate       FLOAT           NOT NULL,
    started_at          TIMESTAMPTZ     NOT NULL,
    ended_at            TIMESTAMPTZ     NOT NULL,
    payment_frequency   VARCHAR(20)     NOT NULL DEFAULT 'weekly',
    total_payments      INT             NOT NULL,
    status              VARCHAR(20)     NOT NULL DEFAULT 'active',

    PRIMARY KEY (id)
);

CREATE TABLE loan_schedules (
    id                  VARCHAR(36)     NOT NULL,
    loan_id             VARCHAR(36)     NOT NULL,
    seq                 INT             NOT NULL,
    due_date            TIMESTAMPTZ     NOT NULL,
    amount_due          JSONB           NOT NULL,
    status              VARCHAR(20)     NOT NULL DEFAULT 'unpaid',

    PRIMARY KEY (id),
    CONSTRAINT fk_loan_id
        FOREIGN KEY(loan_id) 
	    REFERENCES loans(id)
        ON DELETE CASCADE
);