# go-credit

POC for test purposes.

CRUD for account

CRUD for account_balance

## Database

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

## K8 local

Add in hosts file /etc/hosts the lines below

    127.0.0.1   account.domain.com

or

Add -host header in PostMan

## AWS

Create a public apigw

