NODES = 3
DEST =  $(CURDIR)/conf
PASS = $(CURDIR)/../eth/pwd.txt
IPBASE = node
IPADD = 0

conf:
	rm -rf conf
	./../eth/scripts/build-eth-conf.sh $(NODES) $(DEST) $(PASS)
	./scripts/build-raft-conf.sh $(NODES) $(IPBASE) $(IPADD)
	./scripts/edit-toml.sh $(NODES) $(IPBASE) $(IPADD)

compile: ;

.PHONY: conf compile
