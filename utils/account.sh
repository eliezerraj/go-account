#!/bin/bash

echo load ACCOUNT data

var_person=0
genPerson(){
    var_person=$(($RANDOM%($max-$min+1)+$min))
}

# --------------------Load 1 per 1-------------------------
domain=https://go-account.architecture.caradhras.io/add

TOKEN=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6ImFkbWluIiwic2NvcGUiOlsiYWRtaW4iXSwiZXhwIjoxNzA4MDIyODc2fQ.jFrkyd7emiDfz6s_T7UCJ2lsJLOHThUi0bbgumBP6Jg

startid=1

for (( x=0; x<10000; x++ ))
do
    idx=$((startid + x))
    echo curl -X POST $domain -H 'Content-Type: application/json' -H "Authorization: $TOKEN" -d '{"account_id":"ACC-'$idx'","person_id": "P-'$idx'","tenant_id": "TENANT-1"}'
   curl -X POST $domain -H 'Content-Type: application/json' -H "Authorization: $TOKEN" -d '{"account_id":"ACC-'$idx'","person_id": "P-'$idx'","tenant_id": "TENANT-1"}'
done

# --------------------Load n per 1-------------------------
domain=https://go-account.architecture.caradhras.io/add

min=50000
max=51000

startid=50000

for (( x=0; x<=10000; x++ ))
do
    idx=$((startid + x))
    genPerson
    echo curl -X POST $domain -H 'Content-Type: application/json' -H "Authorization: $TOKEN" -d '{"account_id":"ACC-'$idx'","person_id": "P-'$var_person'","tenant_id": "TENANT-1"}'
    curl -X POST $domain -H 'Content-Type: application/json' -H "Authorization: $TOKEN" -d '{"account_id":"ACC-'$idx'","person_id": "P-'$var_person'","tenant_id": "TENANT-1"}'
done