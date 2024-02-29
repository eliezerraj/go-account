truncate	card cascade;
truncate 	payment cascade;
truncate       transfer_moviment cascade;
truncate       account_statement_fee cascade;
truncate       account_statement cascade;
truncate       ACCOUNT_BALANCE cascade;
truncate       ACCOUNT cascade; 
 
drop	table	card cascade;   
drop	table 	payment cascade;   
drop	table   transfer_moviment;
drop	table   account_statement_fee;
drop	table   account_statement cascade;
drop	table   ACCOUNT_BALANCE cascade;
drop	table   ACCOUNT cascade;     

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE ACCOUNT (
            id              SERIAL PRIMARY KEY,
            account_id      varchar(200) UNIQUE NULL,
            person_id       varchar(200) NULL,
            create_at       timestamptz NULL,
            update_at       timestamptz NULL,
            tenant_id       varchar(200) null,
            user_last_update	varchar(200) NULL);

CREATE TABLE ACCOUNT_BALANCE (
            id              SERIAL PRIMARY KEY,
            fk_account_id   integer REFERENCES account(id),
            currency        varchar(10) NULL,   
            amount          float8 NULL,
            create_at       timestamptz NULL,
            update_at       timestamptz NULL,
            tenant_id       varchar(200) null,
            user_last_update	varchar(200) NULL,
            transaction_id	uuid null);

CREATE TABLE account_statement (
            id              SERIAL PRIMARY KEY,
            fk_account_id   integer REFERENCES account(id),
            type_charge     varchar(200) NULL,
            charged_at      timestamptz NULL,
            currency        varchar(10) NULL,   
            amount          float8 NULL,
            tenant_id       varchar(200) NULL
        );

CREATE TABLE account_statement_fee (
        id              SERIAL PRIMARY KEY,
        fk_account_statement_id   integer REFERENCES account_statement(id),
        charged_at      timestamptz NULL,
        type_fee        varchar(200) NULL,
        value_fee       float8 NULL,
        currency        varchar(10) NULL,   
        amount          float8 NULL,
        tenant_id       varchar(200) NULL
    );

CREATE TABLE transfer_moviment (
        id                  SERIAL PRIMARY KEY,
        fk_account_id_from  integer REFERENCES account(id),
        fk_account_id_to    integer REFERENCES account(id),
        type_charge         varchar(200) NULL,
        transfer_at         timestamptz NULL,
        currency            varchar(10) NULL,   
        amount              float8 NULL,
        status              varchar(200) NULL
    );

    CREATE TABLE payment (
        id                  SERIAL PRIMARY KEY,
        fk_card_id          integer REFERENCES card(id),
        card_number         varchar(200) NULL,
        fk_terminal_id      integer REFERENCES terminal(id),
        terminal_name       varchar(200) NULL,
        card_type           varchar(200) NULL,
        card_model          varchar(200) NULL,
        payment_at          timestamptz NULL,
        mcc                 varchar(10) NULL,
        status              varchar(200) NULL,
        currency            varchar(10) NULL,   
        amount              float8 NULL,
        create_at           timestamptz NULL,
        update_at           timestamptz NULL,
        tenant_id           varchar(200) NULL
    );

    CREATE TABLE card (
        id                  SERIAL PRIMARY KEY,
        fk_account_id       integer REFERENCES account(id),
        card_number         varchar(200) NULL,
        card_type           varchar(200) NULL,
        card_model           varchar(200) NULL,
        card_pin            varchar(200) NULL,
        status              varchar(200) NULL,
        expire_at           timestamptz NULL,
        create_at           timestamptz NULL,
        update_at           timestamptz NULL,
        tenant_id           varchar(200) NULL
    );
   
    CREATE TABLE terminal (
        id                  SERIAL PRIMARY KEY,
        terminal_name       varchar(200) NULL,
        coord_x             float8 NULL,
        coord_y             float8 NULL,
        status              varchar(200) NULL,
        create_at           timestamptz NULL,
        update_at           timestamptz NULL
    );