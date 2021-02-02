include .env
DOCO := docker-compose
.PHONY: shell
shell: up; $(DOCO) exec arangodb arangosh --server.password $(DB_PASWORD)

.PHONY: up
up: ; $(DOCO) up -d

.PHONY: down
down: ; $(DOCO) down

import: up ; ./build-graph.sh

