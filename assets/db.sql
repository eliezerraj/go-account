truncate       person cascade; 
truncate       account cascade; 
truncate       account_balance cascade;
truncate       account_statement cascade;
truncate       account_statement_fee cascade;
truncate       card cascade;
truncate       payment cascade;
truncate       terminal cascade;
truncate       transfer_moviment cascade;
truncate       fraud_dataset_view cascade;

drop    table  		person cascade; 
drop    table  		account cascade; 
drop    table       account_balance cascade;
drop    table       account_statement cascade;
drop    table       account_statement_fee cascade;
drop    table       card cascade;
drop    table       payment cascade;
drop    table       terminal cascade;
drop    table       transfer_moviment cascade;
drop    table       fraud_dataset_view cascade;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE public.person (
	id 	serial4		NOT NULL,
	person_id 		varchar(200) NULL,
	person_name 	varchar(200) NULL,
	age 			int NULL,
	profession		int NULL,
	education_level	int NULL,
	salary_level	int NULL,
	gender			varchar(1) NULL,
	create_at 		timestamptz NULL,
	update_at 		timestamptz NULL,
	fk_person_id 	int NULL,
	tenant_id 		varchar(200) NULL,
	CONSTRAINT person_id_key UNIQUE (person_id),
	CONSTRAINT person_pkey PRIMARY KEY (id),
	FOREIGN KEY (fk_person_id) REFERENCES public.person(id)
);

CREATE TABLE public.account (
	id serial4 NOT NULL,
	account_id varchar(200) NULL,
	person_id varchar(200) NULL,
	create_at timestamptz NULL,
	update_at timestamptz NULL,
	tenant_id varchar(200) NULL,
	user_last_update varchar(200) NULL,
	CONSTRAINT account_account_id_key UNIQUE (account_id),
	CONSTRAINT account_pkey PRIMARY KEY (id)
);

CREATE TABLE public.account_balance (
	id serial4 NOT NULL,
	fk_account_id int4 NULL,
	currency varchar(10) NULL,
	amount float8 NULL,
	create_at timestamptz NULL,
	update_at timestamptz NULL,
	tenant_id varchar(200) NULL,
	user_last_update varchar(200) NULL,
	transaction_id uuid NULL,
	CONSTRAINT account_balance_pkey PRIMARY KEY (id)
);

CREATE TABLE public.account_statement (
	id serial4 NOT NULL,
	fk_account_id int4 NULL,
	type_charge varchar(200) NULL,
	charged_at timestamptz NULL,
	currency varchar(10) NULL,
	amount float8 NULL,
	tenant_id varchar(200) NULL,
	CONSTRAINT account_statement_pkey PRIMARY KEY (id)
);

CREATE TABLE public.account_statement_fee (
	id serial4 NOT NULL,
	fk_account_statement_id int4 NULL,
	charged_at timestamptz NULL,
	type_fee varchar(200) NULL,
	value_fee float8 NULL,
	currency varchar(10) NULL,
	amount float8 NULL,
	tenant_id varchar(200) NULL,
	CONSTRAINT account_statement_fee_pkey PRIMARY KEY (id)
);

ALTER TABLE public.account_statement_fee 
ADD CONSTRAINT account_statement_fee_fk_account_statement_id_fkey 
FOREIGN KEY (fk_account_statement_id) REFERENCES public.account_statement(id);

CREATE TABLE public.card (
	id serial4 NOT NULL,
	fk_account_id int4 NULL,
	card_number varchar(200) NULL,
	card_type varchar(200) NULL,
	card_model varchar(200) NULL,
	card_pin varchar(200) NULL,
	status varchar(200) NULL,
	expire_at timestamptz NULL,
	create_at timestamptz NULL,
	update_at timestamptz NULL,
	tenant_id varchar(200) NULL,
	CONSTRAINT card_pkey PRIMARY KEY (id)
);
CREATE INDEX card_idx ON public.card USING btree (card_number);

ALTER TABLE public.card 
ADD CONSTRAINT card_fk_account_id_fkey 
FOREIGN KEY (fk_account_id) REFERENCES public.account(id);

CREATE TABLE public.terminal (
	id serial4 NOT NULL,
	terminal_name varchar(200) NULL,
	coord_x float8 NULL,
	coord_y float8 NULL,
	status varchar(200) NULL,
	create_at timestamptz NULL,
	update_at timestamptz NULL,
	CONSTRAINT terminal_pkey PRIMARY KEY (id)
);

CREATE TABLE public.payment (
	id serial4 NOT NULL,
	fk_card_id int4 NULL,
	card_number varchar(200) NULL,
	fk_terminal_id int4 NULL,
	terminal_name varchar(200) NULL,
	card_type varchar(200) NULL,
	card_model varchar(200) NULL,
	payment_at timestamptz NULL,
	mcc varchar(10) NULL,
	status varchar(200) NULL,
	currency varchar(10) NULL,
	amount float8 NULL,
	create_at timestamptz NULL,
	update_at timestamptz NULL,
	tenant_id varchar(200) NULL,
	fraud float8 NULL,
	CONSTRAINT payment_pkey PRIMARY KEY (id)
);
CREATE INDEX payment_idx ON public.payment USING btree (card_number);

ALTER TABLE public.payment ADD CONSTRAINT payment_fk_card_id_fkey FOREIGN KEY (fk_card_id) REFERENCES public.card(id);
ALTER TABLE public.payment ADD CONSTRAINT payment_fk_terminal_id_fkey FOREIGN KEY (fk_terminal_id) REFERENCES public.terminal(id);


CREATE TABLE public.transfer_moviment (
	id serial4 NOT NULL,
	fk_account_id_from int4 NULL,
	fk_account_id_to int4 NULL,
	type_charge varchar(200) NULL,
	transfer_at timestamptz NULL,
	currency varchar(10) NULL,
	amount float8 NULL,
	status varchar(200) NULL,
	CONSTRAINT transfer_moviment_pkey PRIMARY KEY (id)
);

ALTER TABLE public.transfer_moviment 
ADD CONSTRAINT transfer_moviment_fk_account_id_from_fkey 
FOREIGN KEY (fk_account_id_from) REFERENCES public.account(id);
ALTER TABLE public.transfer_moviment 
ADD CONSTRAINT transfer_moviment_fk_account_id_to_fkey 
FOREIGN KEY (fk_account_id_to) REFERENCES public.account(id);


-- public.fraud_dataset_view source

CREATE OR REPLACE VIEW public.fraud_dataset_view
AS SELECT row_number() OVER (ORDER BY p.payment_at) AS id,
    p.fk_card_id,
    p.card_number,
    p.terminal_name,
    t.coord_x,
    t.coord_y,
    p.card_type,
    p.card_model,
    p.payment_at,
    p.mcc,
    p.amount,
        CASE
            WHEN p.payment_at::time without time zone < '08:00:00'::time without time zone THEN 'night'::text
            WHEN p.payment_at::time without time zone >= '20:00:00'::time without time zone THEN 'night'::text
            ELSE 'day'::text
        END AS night_day,
        CASE
            WHEN p.payment_at::time without time zone < '08:00:00'::time without time zone THEN '1'::text
            WHEN p.payment_at::time without time zone >= '20:00:00'::time without time zone THEN '1'::text
            ELSE '0'::text
        END AS ic_night_day,
        CASE
            WHEN EXTRACT(dow FROM p.payment_at) = ANY (ARRAY[0::numeric, 6::numeric]) THEN 'wkend'::text
            ELSE 'wkday'::text
        END AS wkend_wkday,
        CASE
            WHEN EXTRACT(dow FROM p.payment_at) = ANY (ARRAY[0::numeric, 6::numeric]) THEN '1'::text
            ELSE '0'::text
        END AS ic_wkend_wkday,
    EXTRACT(doy FROM p.payment_at) AS day_of_year,
    ( SELECT count(*) AS tx_1d
           FROM payment p1
          WHERE p1.card_number::text = p.card_number::text AND p1.payment_at::date = p.payment_at::date
          GROUP BY p1.card_number, (p1.payment_at::date)) AS tx_1d,
    ( SELECT to_char(avg(p1.amount), 'FM999999999.00'::text) AS avg_1d
           FROM payment p1
          WHERE p1.card_number::text = p.card_number::text AND p1.payment_at::date = p.payment_at::date
          GROUP BY p1.card_number, (p1.payment_at::date)) AS avg_1d,
    ( SELECT count(*) AS tx_7d
           FROM payment p1
          WHERE p1.card_number::text = p.card_number::text AND p1.payment_at::date >= (p.payment_at::date - '6 days'::interval) AND p1.payment_at::date <= p.payment_at::date
          GROUP BY p1.card_number) AS tx_7d,
    ( SELECT to_char(avg(p1.amount), 'FM999999999.00'::text) AS avg_7d
           FROM payment p1
          WHERE p1.card_number::text = p.card_number::text AND p1.payment_at::date >= (p.payment_at::date - '6 days'::interval) AND p1.payment_at::date <= p.payment_at::date
          GROUP BY p1.card_number) AS avg_7d,
    ( SELECT count(*) AS tx_30d
           FROM payment p1
          WHERE p1.card_number::text = p.card_number::text AND p1.payment_at::date >= (p.payment_at::date - '31 days'::interval) AND p1.payment_at::date <= p.payment_at::date
          GROUP BY p1.card_number) AS tx_30d,
    ( SELECT to_char(avg(p1.amount), 'FM999999999.00'::text) AS avg_30d
           FROM payment p1
          WHERE p1.card_number::text = p.card_number::text AND p1.payment_at::date >= (p.payment_at::date - '31 days'::interval) AND p1.payment_at::date <= p.payment_at::date
          GROUP BY p1.card_number) AS avg_30d,
    to_char(COALESCE(EXTRACT(epoch FROM p.payment_at - lag(p.payment_at) OVER (ORDER BY p.payment_at DESC)), 0::numeric) * '-1'::integer::numeric, 'FM999999999'::text) AS time_btw_tx,
    to_char(COALESCE(( SELECT EXTRACT(epoch FROM p.payment_at - p1.payment_at) AS "extract"
           FROM payment p1
          WHERE p1.card_number::text = p.card_number::text AND p1.payment_at < p.payment_at AND p1.payment_at::date = p.payment_at::date
          ORDER BY p1.payment_at DESC
         LIMIT 1), 0::numeric), 'FM999999999'::text) AS time_btw_cc_tx,
    to_char(p.fraud, 'FM999999999'::text) AS fraud
   FROM payment p,
    terminal t
  WHERE p.fk_terminal_id = t.id AND p.fk_card_id < 1000
  ORDER BY p.payment_at;