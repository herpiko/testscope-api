include .env
export $(shell sed 's/=.*//' .env)

prep:
	docker kill $$(docker ps -q) || true
	docker network rm ${PROJECT_NAME}_default; true
	docker network create -d bridge ${PROJECT_NAME}_default --opt external=true; true
	docker-compose -p ${PROJECT_NAME} up -d --force-recreate proxy
	docker-compose -p ${PROJECT_NAME} up -d --force-recreate testdb
	docker-compose -p ${PROJECT_NAME} up -d --force-recreate testminio
	docker run --network ${PROJECT_NAME}_default willwill/wait-for-it testdb:5432 -- echo "database is up"
	docker-compose -p ${PROJECT_NAME} run testdb createdb -h testdb -U ${DB_USER} -w ${DB_NAME} || true

proxy:
	docker-compose -p ${PROJECT_NAME} up -d --force-recreate proxy

dbshell:
	docker exec -ti ${PROJECT_NAME}-testdb-1 psql -U testdb -d testdb

run:
	GOWORK=off go run .

test:
	go test -coverprofile=coverage.html

cov:
	go tool cover -html=coverage.html

build-docker:
	docker build --platform linux/amd64 -t herpiko/${PROJECT_NAME}-api:latest .

deploy: build-docker
	docker save testscopeio-api:latest > testscopeio-api-latest.img
	scp testscopeio-api-latest.img user@foobar:~/api.img
	ssh user@foobar 'docker load < api.img'
	ssh user@foobar 'docker-compose up -d --force-recreate api'

dockertest:
	docker run -ti \
	--network ${PROJECT_NAME}_default \
	-e PROJECT_NAME=${PROJECT_NAME} \
	-e APP_NAME=${APP_NAME} \
	-e DB_HOST='testdb' \
	-e DB_USER=${DB_USER} \
	-e DB_PASS=${DB_PASS} \
	-e DB_NAME=${DB_NAME} \
	-e FIREBASE_ACCOUNT_KEY_PATH=${FIREBASE_ACCOUNT_KEY_PATH} \
	-e FIREBASE_PROJECT_ID=${FIREBASE_PROJECT_ID} \
	-e XENDIT_API_KEY=${XENDIT_API_KEY} \
	-e XENDIT_API_SECRET=${XENDIT_API_SECRET} \
	-e XENDIT_API_PUB_KEY=${XENDIT_API_PUB_KEY} \
	-e XENDIT_CALLBACK_TOKEN=${XENDIT_CALLBACK_TOKEN} \
	-e SENDINDBLUE_API_KEY=${SENDINDBLUE_API_KEY} \
	${PROJECT_NAME}-api /app/scripts/test.sh

restoreprod:
	ssh user@foobar 'docker exec -t user_db_1 pg_dumpall -c -U db > dump.sql' || true
	scp user@foobar:~/dump.sql .
	docker-compose stop testdb
	docker-compose -p ${PROJECT_NAME} run testdb dropdb -h template1 -U ${DB_USER} -w ${DB_NAME} || true
	docker-compose -p ${PROJECT_NAME} run testdb dropdb -h testdb -U ${DB_USER} -w ${DB_NAME} || true
	docker-compose -p ${PROJECT_NAME} run testdb createdb -h testdb -U ${DB_USER} -w ${DB_NAME} || true
	sed -i'' -e 's/testscopeio/testdb/g' dump.sql
	#sed -i 's/testscopeio/testdb/g' dump.sql
	cat dump.sql | docker exec -i $$(docker ps | grep testscope | grep testdb | cut -d ' ' -f 1) psql -U testdb

db-tunnel:
	ssh -vvv -L 5433:foobar:5432 user@testscope.io -p 22 -N
