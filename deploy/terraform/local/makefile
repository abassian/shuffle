NODES = 1
CONSENSUS = solo
VERSION = latest
CONF = $(CURDIR)/../../conf/$(CONSENSUS)/conf

apply:
	rm -f ips.dat && \
	terraform apply -auto-approve \
		-var user=$(shell id -u) \
		-var consensus=$(CONSENSUS) \
		-var nodes=$(NODES) \
		-var version=$(VERSION) \
		-var conf=$(CONF)
		
destroy:
	terraform destroy -auto-approve

.PHONY: apply destroy