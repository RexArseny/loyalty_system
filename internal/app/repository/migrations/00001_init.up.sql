START TRANSACTION;

CREATE TABLE users (
	user_id uuid NOT NULL,
	login text NOT NULL,
	hash text NOT NULL,
	salt text NOT NULL,
	CONSTRAINT users_pk PRIMARY KEY (user_id),
	CONSTRAINT users_unique UNIQUE (login)
);

CREATE TABLE orders (
  order_id text NOT NULL,
  status text NOT NULL,
  accrual double precision,
  uploaded_at timestamp with time zone NOT NULL,
	user_id uuid NOT NULL,
	CONSTRAINT orders_pk PRIMARY KEY (order_id)
);

CREATE TABLE balances (
	user_id uuid NOT NULL,
  balance double precision NOT NULL,
  withdrawn double precision NOT NULL,
	CONSTRAINT balances_pk PRIMARY KEY (user_id)
);

CREATE TABLE withdrawals (
	user_id uuid NOT NULL,
  order_id text NOT NULL,
	sum double precision NOT NULL,
	processed_at timestamp with time zone NOT NULL
);

COMMIT;