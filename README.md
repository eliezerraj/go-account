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
            request_id varchar NULL,
            jwt_id varchar NULL,
            CONSTRAINT account_balance_pkey PRIMARY KEY (id)
        );

        CREATE TABLE audit_account_balance (
            id serial4 			NOT NULL,
            request_id 			varchar NULL,
            jwt_id 				varchar NULL,
            user_session 		varchar NOT NULL,
            create_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,            
            account_balance_id 	INT NOT NULL,
            currency 			varchar NULL,
            old_amount 			float8 NOT NULL,
            new_amount 			float8 NOT null
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

## 

cat xxx.key | base64 -w 0

##

        CREATE OR REPLACE FUNCTION func_audit_account_balance() RETURNS trigger as $$
        DECLARE
            sess_user text;
        begin
            BEGIN
                sess_user := (SELECT session_user from session_user);
            EXCEPTION WHEN undefined_table THEN
                sess_user := 'unknown_user';
            END;
            INSERT INTO audit_account_balance (	request_id, 
                                                jwt_id,
                                                user_session,
                                                account_balance_id, 
                                                currency, 
                                                old_amount, 
                                                new_amount)
            VALUES (OLD.request_id,
                    OLD.jwt_id,
                    sess_user,
                    OLD.id,
                    OLD.currency, 
                    OLD.amount, 
                    NEW.amount);
            
            RETURN NEW;
        END;
        $$ LANGUAGE plpgsql;


        CREATE TRIGGER trg_audit_account_balance
        AFTER UPDATE OF amount ON account_balance
        FOR EACH ROW
        EXECUTE FUNCTION func_audit_account_balance();