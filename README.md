# go-credit

POC for test purposes.

    CRUD for account

# Run
    go run . >& /mnt/c/Eliezer/log/go-account.log

## Database

    See repo https://github.com/eliezerraj/go-account-migration-worker.git

## Endpoints

+ GET /header

+ GET /info

+ POST /add

        {
            "account_id": "ACC-1.1",
            "person_id": "P-1",
            "tenant_id": "TENANT-1"
        }

+ GET /get/ACC-003

+ GET /list/P-002

+ POST /update/ACC-003

        {
            "person_id": "P-002",
            "tenant_id": "TENANT-001"
        }

+ DELETE /delete/ACC-001

+ GET /movimentAccountBalance/ACC-100

+ GET /accountBalance/ACC-20

## K8 local

    Add in hosts file /etc/hosts the lines below

    127.0.0.1   account.domain.com

    or

    Add -host header in PostMan

