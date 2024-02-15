#!/bin/bash

echo load ACCOUNT data

var_person=0
genPerson(){
    var_person=$(($RANDOM%($max-$min+1)+$min))
}

# --------------------Load 1 per 1-------------------------
domain=http://localhost:5000/add
startid=1

for (( x=0; x<=10; x++ ))
do
    idx=$((startid + x))
    echo curl -X POST $domain -H 'Content-Type: application/json' -d '{"account_id":"ACC-'$idx'","person_id": "P-'$idx'","tenant_id": "TENANT-1"}'
    curl -X POST $domain -H 'Content-Type: application/json' -d '{"account_id":"ACC-'$idx'","person_id": "P-'$idx'","tenant_id": "TENANT-1"}'
done

# --------------------Load n per 1-------------------------
domain=http://localhost:5000/add

min=500
max=550

startid=500

for (( x=0; x<=10; x++ ))
do
    idx=$((startid + x))
    genPerson
    echo curl -X POST $domain -H 'Content-Type: application/json' -d '{"account_id":"ACC-'$idx'","person_id": "P-'$var_person'","tenant_id": "TENANT-1"}'
    curl -X POST $domain -H 'Content-Type: application/json' -d '{"account_id":"ACC-'$idx'","person_id": "P-'$var_person'","tenant_id": "TENANT-1"}'
done