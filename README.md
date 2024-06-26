# go-credit

POC for test purposes.

CRUD for account

CRUD for account_balance

## Database

        CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

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

## Endpoints

+ POST /add

        {
            "account_id": "ACC-1.1",
            "person_id": "P-1",
            "tenant_id": "TENANT-1"
        }

+ GET /get/ACC-003

        curl account.domain.com/get/ACC-001 | jq

+ GET /header

+ GET /list/P-002

        curl account.domain.com/list/P-003 | jq

+ POST /update/ACC-003

        {
            "person_id": "P-002",
            "tenant_id": "TENANT-001"
        }

+ DELETE /delete/ACC-001

+ GET /movimentBalanceAccount/ACC-100

+ GET /fundBalanceAccount/ACC-20

+ POST /transferFund

        {
            "account_id_from": "ACC-500",
            "fk_account_id_from": 11,
            "account_id_to": "ACC-600",
            "fk_account_id_TO": 12,
            "type_charge": "TRANSFER1",
            "currency": "BRL",
            "amount": 1.00
        }


## K8 local

Add in hosts file /etc/hosts the lines below

    127.0.0.1   account.domain.com

or

Add -host header in PostMan

## AWS

Create a public apigw

