SHELL := bash

# Define environment variables
export AUTH_TOKEN=

export URL_POST=https://go-global.architecture.caradhras.io/account/add

# Default target
all: env load

# Show environment variables
env:
	@echo "Current Environment Variables:"
	@echo "AUTH_TOKEN=$(AUTH_TOKEN)"
	@echo "URL_POST=$(URL_POST)"
	@echo "---------------------------------------------------------"
	
load:
	@echo "Run Load Account ..."
	
	@for ((i=1; i<=1; i++)); do \
		ACC_ID=$$((4999 + i)); \
		echo "Posting iteration $$ACC_ID... {"account_id": "ACC-$$ACC_ID","person_id":"P-$$ACC_ID"}"; \
		curl -X POST $(URL_POST) \
		    --header "Content-Type: application/json" \
			--header "Authorization: $(AUTH_TOKEN)" \
		    --data '{"account_id":"ACC-'$$ACC_ID'","person_id":"P-'$$ACC_ID'"}'; \
		echo ""; \
	done

.PHONY: all env load