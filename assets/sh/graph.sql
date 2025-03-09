
# Merchant
SELECT json_agg(
           json_build_object(
               'id', id,
               'label','merchant',
               'properties',json_build_object(	'id', id,
               									'name', id)
           )
       ) AS result
FROM (select distinct(mcc) as id from payment p ) t

## Terminal
SELECT json_agg(
           json_build_object(
               'id', terminal_name,
               'label','terminal',
               'properties',json_build_object(	'id', terminal_name,
               									'name', terminal_name,
               									'status', status,
               									'coord_x',coord_x,
               									'coord_y',coord_y)
           )
       ) AS result
FROM terminal t 

## Person
SELECT json_agg(
           json_build_object(
               'id', person_id,
               'label','person',
               'properties',json_build_object(	'id', person_id,
               									'name', person_name,
               									'person_id', person_id,
               									'age', age,
               									'profession',profession,
               									'education_level',education_level)
           )
       ) AS result
FROM person

## Account
SELECT json_agg(
           json_build_object(
               'id', account_id,
               'label','account',
               'properties',json_build_object(	'id', account_id,
               									'name', account_id,
               									'person', person_id)
           )
       ) AS result
FROM account a 

## Card
SELECT json_agg(
           json_build_object(
               'id', card_number,
               'label','card',
               'properties',json_build_object(	'id', card_number,
               									'card_number', card_number,
               									'account', account_id,
               									'type', card_type,
               									'model', card_model,
               									'status', status)
           )
       ) AS result
FROM (select 	c.id,
				c.card_number,
				a.account_id, 
				c.card_type, 
				c.card_model, 
				c.status 
		from 	card c, 
				account a 
		where c.fk_account_id = a.id ) a 

## Account
SELECT json_agg(
           json_build_object(
               'id', account_id,
               'label','account',
               'properties',json_build_object(	'id', account_id,
               									'name', account_id,
               									'person', person_id)
           )
       ) AS result
FROM account a 

## Payment
SELECT json_agg(
           json_build_object(
               'id', id,
               'label','payment',
               'properties',json_build_object(	'id', id,
               									'currency', currency,
               									'amount',amount,
               									'type', card_type,
               									'model',card_model,
               									'payment_at' ,payment_at)
           )
       ) AS result
FROM (select 	p.id,
				p.currency, 
				p.amount,
				p.card_type,
				p.card_model,
				p.payment_at,
				p.mcc, 
				t.terminal_name,
				c.card_number 
		from 	payment p,
				terminal t,
				card c 
		where	p.fk_terminal_id = t.id
		and 	p.fk_card_id = c.id ) a 

========================================================
		
## Edge Person / Account
SELECT json_agg(
           json_build_object(
               'from_id', person_id,
               'to_id', account_id,
               'label','has',
               'properties',json_build_object('id', person_id || '-' || account_id || '-ACC-PER')
           )
       ) AS result
FROM account a 

## Edge Account / Card
SELECT json_agg(
           json_build_object(
               'from_id', account_id,
               'to_id', card_number,
               'label','issued',
               'properties',json_build_object('id', account_id || '-' || card_number || '-ACC-CARD')
           )
       ) AS result
FROM (select 	c.card_number,
				a.account_id, 
				c.card_type, 
				c.card_model, 
				c.status 
		from 	card c, 
				account a 
		where c.fk_account_id = a.id) a 
		
## Edge Card / Payment
SELECT json_agg(
           json_build_object(
               'from_id', card_number,
               'to_id', id,
               'label','payed',
               'properties',json_build_object('id', id || '-' || card_number || '-PAY-CARD')
           )
       ) AS result
FROM (select 	p.id,
				p.currency, 
				p.amount,
				p.card_type,
				p.card_model,
				p.payment_at,
				p.mcc, 
				t.terminal_name,
				c.card_number 
		from 	payment p,
				terminal t,
				card c 
		where	p.fk_terminal_id = t.id
		and 	p.fk_card_id = c.id ) a 
		
		
## Edge Payment / MCC
SELECT json_agg(
           json_build_object(
               'from_id', id,
               'to_id', mcc,
               'label','payed_at',
               'properties',json_build_object('id', id || '-' || mcc || '-PAY-MCC')
           )
       ) AS result
FROM (select 	p.id,
				p.currency, 
				p.amount,
				p.card_type,
				p.card_model,
				p.payment_at,
				p.mcc, 
				t.terminal_name,
				c.card_number 
		from 	payment p,
				terminal t,
				card c 
		where	p.fk_terminal_id = t.id
		and 	p.fk_card_id = c.id ) a 
	
## Edge Payment / Terminal
SELECT json_agg(
           json_build_object(
               'from_id', id,
               'to_id', terminal_name,
               'label','payed_at',
               'properties',json_build_object('id', id || '-' || terminal_name || '-PAY-TERMINAL')
           )
       ) AS result
FROM (select 	p.id,
				p.currency, 
				p.amount,
				p.card_type,
				p.card_model,
				p.payment_at,
				p.mcc, 
				t.terminal_name,
				c.card_number 
		from 	payment p,
				terminal t,
				card c 
		where	p.fk_terminal_id = t.id
		and 	p.fk_card_id = c.id ) a 