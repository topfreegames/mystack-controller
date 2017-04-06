# mystack-controller api
# https://github.com/topfreegames/mystack-controller
#
# Licensed under the MIT license:
# http://www.opensource.org/licenses/mit-license
# Copyright © 2017 Top Free Games <backend@tfgco.com>

MY_IP=`ifconfig | grep --color=none -Eo 'inet (addr:)?([0-9]*\.){3}[0-9]*' | grep --color=none -Eo '([0-9]*\.){3}[0-9]*' | grep -v '127.0.0.1' | head -n 1`

setup: setup-hooks
	@go get -u github.com/golang/dep...
	@go get -u github.com/jteeuwen/go-bindata/...
	@go get -u github.com/wadey/gocovmerge
	@dep ensure -update

setup-hooks:
	@cd .git/hooks && ln -sf ../../hooks/pre-commit.sh pre-commit

build:
	@mkdir -p bin && go build -o ./bin/mystack-controller main.go

build-docker: cross-build-linux-amd64
	@docker build -t mystack-controller .

cross-build-linux-amd64:
	@env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ./bin/mystack-controller-linux-amd64
	@chmod a+x ./bin/mystack-controller-linux-amd64

assets:
	@go-bindata -o migrations/migrations.go -pkg migrations migrations/*.sql

migrate: assets
	@go run main.go migrate -c ./config/local.yaml

migrate-test: assets
	@go run main.go migrate -c ./config/test.yaml

deps: start-deps wait-for-pg

start-deps:
	@echo "Starting dependencies using HOST IP of ${MY_IP}..."
	@env MY_IP=${MY_IP} docker-compose --project-name mystack up -d
	@sleep 10
	@echo "Dependencies started successfully."

stop-deps:
	@env MY_IP=${MY_IP} docker-compose --project-name mystack down

wait-for-pg:
	@until docker exec mystack_postgres_1 pg_isready; do echo 'Waiting for Postgres...' && sleep 1; done
	@sleep 2

drop:
	@-psql -d postgres -h localhost -p 8585 -U postgres -c "SELECT pg_terminate_backend(pid.pid) FROM pg_stat_activity, (SELECT pid FROM pg_stat_activity where pid <> pg_backend_pid()) pid WHERE datname='mystack';"
	@psql -d postgres -h localhost -p 8585 -U postgres -f scripts/drop.sql > /dev/null
	@echo "Database created successfully!"

drop-test:
	@-psql -d postgres -h localhost -p 8585 -U postgres -c "SELECT pg_terminate_backend(pid.pid) FROM pg_stat_activity, (SELECT pid FROM pg_stat_activity where pid <> pg_backend_pid()) pid WHERE datname='mystack-test';"
	@psql -d postgres -h localhost -p 8585 -U postgres -f scripts/drop-test.sql > /dev/null
	@echo "Test Database created successfully!"

run:
	@go run main.go start -v3 -c ./config/local.yaml

run-full: deps drop migrate run

unit: clear-coverage-profiles unit-run gather-unit-profiles

clear-coverage-profiles:
	@find . -name '*.coverprofile' -delete

unit-run:
	@ginkgo -tags unit -cover -r -randomizeAllSpecs -randomizeSuites -skipMeasurements ${TEST_PACKAGES}

gather-unit-profiles:
	@mkdir -p _build
	@echo "mode: count" > _build/coverage-unit.out
	@bash -c 'for f in $$(find . -name "*.coverprofile"); do tail -n +2 $$f >> _build/coverage-unit.out; done'

integration int: clear-coverage-profiles integration-run gather-integration-profiles

integration-run:
	@ginkgo -tags integration -cover -r -randomizeAllSpecs -randomizeSuites -skipMeasurements ${TEST_PACKAGES}

gather-integration-profiles:
	@mkdir -p _build
	@echo "mode: count" > _build/coverage-integration.out
	@bash -c 'for f in $$(find . -name "*.coverprofile"); do tail -n +2 $$f >> _build/coverage-integration.out; done'

merge-profiles:
	@mkdir -p _build
	@gocovmerge _build/*.out > _build/coverage-all.out

test-coverage-func coverage-func: merge-profiles
	@echo
	@echo "=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-"
	@echo "Functions NOT COVERED by Tests"
	@echo "=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-"
	@go tool cover -func=_build/coverage-all.out | egrep -v "100.0[%]"

test: unit integration test-coverage-func
