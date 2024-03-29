ENV = local # local or aws
VERSION = 0.2.1 # shuffle Docker Image
CONSENSUS = solo # solo, huron, or raft
NODES = 1
VALIDATORS = 1
POA = true
IPBASE = node # use 10.0.2. for aws 
IPADD = 0 # use 10 for aws
PREBUILT = "" # prebuilt configurations
CONSENSUSPORT = 1337

up: conf compile start

build-docker-image:
	$(MAKE) -C build docker

build-ami:
	$(MAKE) -C build ami

check: 
	scripts/check_docker_running.sh	

conf:
	$(MAKE) -C conf/$(CONSENSUS) conf NODES=$(NODES) IPBASE=$(IPBASE) IPADD=$(IPADD) VALIDATORS=$(VALIDATORS) POA=$(POA) CONSENSUSPORT=$(CONSENSUSPORT)

compile:
	$(MAKE) -C conf/$(CONSENSUS) compile

prebuilt:
	$(MAKE) -C prebuilt prebuilt PREBUILT="$(PREBUILT)"

start: check startnocheck

startnocheck: 
	$(MAKE) -C terraform/$(ENV) apply NODES=$(NODES) CONSENSUS=$(CONSENSUS) VERSION=$(VERSION)

startnoerr: 
	-$(MAKE) -C terraform/$(ENV) apply NODES=$(NODES) CONSENSUS=$(CONSENSUS) VERSION=$(VERSION)

stop:
	$(MAKE) -C terraform/$(ENV) destroy

checktools: solc-exists npm-exists docker-exists
solc-exists: ; @which solc > /dev/null
npm-exists: ; @which npm > /dev/null
docker-exists: ; @which docker > /dev/null



.PHONY: up build-docker build-ami check conf prebuilt start startnoerr startnocheck stop checktools solc-exists npm-exists docker-exists compile
