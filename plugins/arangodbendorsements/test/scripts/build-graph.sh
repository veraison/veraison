#!/bin/bash
# Copyright 2021 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

. .env
COMPOSE="docker-compose exec -T arangodb"

ARANGOIMP="$COMPOSE arangoimp"
ARANGOSHELL="$COMPOSE arangosh --server.password ${DB_PASSWORD} --server.database ${DB_NAME}"

BASEDIR="/import"

for f in import/*.json
do
	bn="$(basename $f .json)"
	fn="${BASEDIR}/$bn.json"

	echo "Importing $bn"
	echo "======================================="

	if [[ $bn == edge_* ]]
	then
		$ARANGOIMP \
			--server.password ${DB_PASSWORD} \
			--server.database ${DB_NAME} \
			--create-database true \
			--file $fn \
			--collection=$bn \
			--create-collection true \
			--create-collection-type edge \
			--type=json \
			--overwrite true
	else
		$ARANGOIMP \
			--server.password ${DB_PASSWORD} \
			--server.database ${DB_NAME} \
			--create-database true \
			--file $fn \
			--collection=$bn \
			--create-collection true \
			--type=json \
			--overwrite true
	fi

	echo
done

$ARANGOSHELL --javascript.execute import/create-graph-from-collections.js
