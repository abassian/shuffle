NODE = node4

deps:
	cd crowd-funding && npm install

deploy:
	$(MAKE) -C ../deploy check  && \
	$(MAKE) -C ../deploy prebuilt PREBUILT=poa && \
	$(MAKE) -C ../deploy startnoerr CONSENSUS=huron NODES=10 && \
	sleep 2 && \
	./scripts/get_running_nodes.sh && \
	./scripts/write_shlc_config.sh node0 
# && \./scripts/init_poa_contract.sh node0

stop:
	$(MAKE) -C ../deploy stop

watch:
	./scripts/watch.sh

crowd-funding-demo:
	./scripts/crowd-funding-demo.sh

add-node:
	./scripts/write_shlc_config.sh node0 && \
	./scripts/nominate_node.sh --from node0 --nominee $(NODE) && \
	./scripts/all_vote_node.sh $(NODE) && \
	./scripts/set_peers.sh $(NODE) && \
	./scripts/node_restart.sh $(NODE) && \
	./scripts/get_running_nodes.sh 

.PHONY: deps deploy stop watch crowd-funding-demo add-node
