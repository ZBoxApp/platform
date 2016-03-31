.PHONY: build package run stop run-client run-server stop-client stop-server restart-server restart-client start-docker clean-dist clean nuke check-style check-unit-tests test dist setup-mac prepare-enteprise zbox zbox-run zbox-dev zbox-prod deploy

# Build Flags
BUILD_NUMBER ?= $(BUILD_NUMBER:)
BUILD_DATE = $(shell date -u)
BUILD_HASH = $(shell git rev-parse HEAD)
# If we don't set the build number it defaults to dev
ifeq ($(BUILD_NUMBER),)
	BUILD_NUMBER := dev
endif

ifeq ($(TRAVIS_BUILD_NUMBER),)
	BUILD_NUMBER := dev
else
	BUILD_NUMBER := $(TRAVIS_BUILD_NUMBER)
endif


BUILD_ENTERPRISE_DIR ?= ../enterprise
BUILD_ENTERPRISE ?= true
BUILD_ENTERPRISE_READY = false
ifneq ($(wildcard $(BUILD_ENTERPRISE_DIR)/.),)
	ifeq ($(BUILD_ENTERPRISE),true)
		BUILD_ENTERPRISE_READY = true
	else
		BUILD_ENTERPRISE_READY = false
	endif
else
	BUILD_ENTERPRISE_READY = false
endif
BUILD_WEBAPP_DIR = ./webapp

# Golang Flags
GOPATH ?= $(GOPATH:)
GOFLAGS ?= $(GOFLAGS:)
GO=$(GOPATH)/bin/godep go
GO_LINKER_FLAGS ?= -ldflags \
				   "-X github.com/mattermost/platform/model.BuildNumber=$(BUILD_NUMBER)\
					-X 'github.com/mattermost/platform/model.BuildDate=$(BUILD_DATE)'\
					-X github.com/mattermost/platform/model.BuildHash=$(BUILD_HASH)\
					-X github.com/mattermost/platform/model.BuildEnterpriseReady=$(BUILD_ENTERPRISE_READY)"

# Output paths
DIST_ROOT=dist
DIST_PATH=$(DIST_ROOT)/zboxnow

# Tests
TESTS=.

DOCKERNAME ?= zboxnow
DOCKER_CONTAINER_NAME ?= zboxnow-dev

all: dist

dist: | zbox test

dist-travis: | travis-init build-container

travis-init:
	@echo Setting up enviroment for travis

	if [ "$(TRAVIS_DB)" = "postgres" ]; then \
		sed -i'.bak' 's|mysql|postgres|g' config/config.json; \
		sed -i'.bak2' 's|zbuser:zbtest@tcp(dockerhost:3306)/zboxnow?charset=utf8mb4,utf8|postgres://zbuser:zbtest@postgres:5432/zboxnow?sslmode=disable\&connect_timeout=10|g' config/config.json; \
	fi

	if [ "$(TRAVIS_DB)" = "mysql" ]; then \
		sed -i'.bak' 's|zbuser:zbtest@tcp(dockerhost:3306)/zboxnow?charset=utf8mb4,utf8|zbuser:zbtest@tcp(mysql:3306)/zboxnow?charset=utf8mb4,utf8|g' config/config.json; \
	fi

build-container:
	@echo Building in container

	cd .. && docker run -e TRAVIS_BUILD_NUMBER=$(TRAVIS_BUILD_NUMBER) -e LANG=en --link zboxnow-mysql:mysql --link zboxnow-postgres:postgres -v `pwd`:/go/src/github.com/mattermost zboxapp/builder:latest

start-docker:
	@echo Starting docker containers

	@if [ $(shell docker ps -a | grep -ci zboxnow-mysql) -eq 0 ]; then \
		echo starting zboxnow-mysql; \
		docker run --name zboxnow-mysql -p 3306:3306 -e MYSQL_ROOT_PASSWORD=zbtest \
		-e MYSQL_USER=zbuser -e MYSQL_PASSWORD=zbtest -e MYSQL_DATABASE=zboxnow -d mysql:5.7 > /dev/null; \
	elif [ $(shell docker ps | grep -ci zboxnow-mysql) -eq 0 ]; then \
		echo restarting zboxnow-mysql; \
		docker start zboxnow-mysql > /dev/null; \
	fi

	@if [ $(shell docker ps -a | grep -ci zboxnow-postgres) -eq 0 ]; then \
		echo starting zboxnow-postgres; \
		docker run --name zboxnow-postgres -p 5432:5432 -e POSTGRES_USER=zbuser -e POSTGRES_PASSWORD=zbtest \
		-d postgres:9.4 > /dev/null; \
		sleep 10; \
	elif [ $(shell docker ps | grep -ci zboxnow-postgres) -eq 0 ]; then \
		echo restarting zboxnow-postgres; \
		docker start zboxnow-postgres > /dev/null; \
		sleep 10; \
	fi

stop-docker:
	@echo Stopping docker containers

	@if [ $(shell docker ps -a | grep -ci zboxnow-mysql) -eq 1 ]; then \
		echo stopping zboxnow-mysql; \
		docker stop zboxnow-mysql > /dev/null; \
	fi

	@if [ $(shell docker ps -a | grep -ci zboxnow-postgres) -eq 1 ]; then \
		echo stopping zboxnow-postgres; \
		docker stop zboxnow-postgres > /dev/null; \
	fi

clean-docker:
	@echo Removing docker containers

	@if [ $(shell docker ps -a | grep -ci zboxnow-mysql) -eq 1 ]; then \
		echo stopping zboxnow-mysql; \
		docker stop zboxnow-mysql > /dev/null; \
		docker rm -v zboxnow-mysql > /dev/null; \
	fi

	@if [ $(shell docker ps -a | grep -ci zboxnow-postgres) -eq 1 ]; then \
		echo stopping zboxnow-postgres; \
		docker stop zboxnow-postgres > /dev/null; \
		docker rm -v zboxnow-postgres > /dev/null; \
	fi

check-style:
	@echo Running GOFMT
	$(eval GOFMT_OUTPUT := $(shell gofmt -d -s api/ model/ store/ utils/ manualtesting/ einterfaces/ mattermost.go 2>&1))
	@echo "$(GOFMT_OUTPUT)"
	@if [ ! "$(GOFMT_OUTPUT)" ]; then \
		echo "gofmt sucess"; \
	else \
		echo "gofmt failure"; \
		exit 1; \
	fi

test:
	sed -i'.bak' 's|"EnableOAuthServiceProvider": true,|"EnableOAuthServiceProvider": false,|g' config/config.json

	$(GO) test $(GOFLAGS) -run=$(TESTS) -test.v -test.timeout=180s ./api || exit 1
	$(GO) test $(GOFLAGS) -run=$(TESTS) -test.v -test.timeout=12s ./model || exit 1
	$(GO) test $(GOFLAGS) -run=$(TESTS) -test.v -test.timeout=120s ./store || exit 1
	$(GO) test $(GOFLAGS) -run=$(TESTS) -test.v -test.timeout=120s ./utils || exit 1
	$(GO) test $(GOFLAGS) -run=$(TESTS) -test.v -test.timeout=120s ./web || exit 1

	mv ./config/config.json.bak ./config/config.json

.prebuild:
	@echo Preparation for running go code
	go get $(GOFLAGS) github.com/tools/godep

	touch $@

prepare-enterprise:
ifeq ($(BUILD_ENTERPRISE_READY),true)
	@echo Enterprise build selected, perparing
	cp $(BUILD_ENTERPRISE_DIR)/imports.go .
endif

build: .prebuild prepare-enterprise
	@echo Building ZBox Now! server

	$(GO) clean $(GOFLAGS) -i ./...
	$(GO) install $(GOFLAGS) $(GO_LINKER_FLAGS) ./...

build-client:
	@echo Building ZBox Now! web app

	cd $(BUILD_WEBAPP_DIR) && $(MAKE) build


package: build build-client
	@ echo Packaging mattermost

	# Remove any old files
	rm -Rf $(DIST_ROOT)

	# Create needed directories
	mkdir -p $(DIST_PATH)/bin
	mkdir -p $(DIST_PATH)/logs

	# Copy binary
	cp $(GOPATH)/bin/platform $(DIST_PATH)/bin

	# Resource directories
	cp -RL config $(DIST_PATH)
	cp -RL fonts $(DIST_PATH)
	cp -RL templates $(DIST_PATH)
	cp -RL i18n $(DIST_PATH)

	# Package webapp
	mkdir -p $(DIST_PATH)/webapp/dist
	cp -RL $(BUILD_WEBAPP_DIR)/dist $(DIST_PATH)/webapp
	mv $(DIST_PATH)/webapp/dist/bundle.js $(DIST_PATH)/webapp/dist/bundle-$(BUILD_NUMBER).js
	sed -i'.bak' 's|bundle.js|bundle-$(BUILD_NUMBER).js|g' $(DIST_PATH)/webapp/dist/root.html
	rm $(DIST_PATH)/webapp/dist/root.html.bak

	# Help files
ifeq ($(BUILD_ENTERPRISE_READY),true)
	cp $(BUILD_ENTERPRISE_DIR)/ENTERPRISE-EDITION-LICENSE.txt $(DIST_PATH)
else
	cp build/MIT-COMPILED-LICENSE.md $(DIST_PATH)
endif
	cp NOTICE.txt $(DIST_PATH)
	cp README.md $(DIST_PATH)

	# Create package
	tar -C dist -czf $(DIST_PATH).tar.gz mattermost

run-server: prepare-enterprise
	@echo Running ZBox Now for development

	mkdir -p $(BUILD_WEBAPP_DIR)/dist/files
	$(GO) run $(GOFLAGS) $(GO_LINKER_FLAGS) *.go &

run-client:
	@echo Running ZBox Now! client for development

	cd $(BUILD_WEBAPP_DIR) && $(MAKE) run

run-client-fullmap:
	@echo Running mattermost client for development with FULL SOURCE MAP

	cd $(BUILD_WEBAPP_DIR) && $(MAKE) run-fullmap

run: run-server run-client

run-fullmap: run-server run-client-fullmap

stop-server:
	@echo Stopping ZBox Now!

	@for PID in $$(ps -ef | grep "[g]o run" | awk '{ print $$2 }'); do \
		echo stopping go $$PID; \
		kill $$PID; \
	done

	@for PID in $$(ps -ef | grep "[g]o-build" | awk '{ print $$2 }'); do \
		echo stopping ZBox Now $$PID; \
		kill $$PID; \
	done

stop-client:
	@echo Stopping ZBox Now client

	cd $(BUILD_WEBAPP_DIR) && $(MAKE) stop


stop: stop-server stop-client

restart-server: | stop-server run-server

restart-client: | stop-client run-client

clean: stop-docker
	@echo Cleaning

	rm -Rf $(DIST_ROOT)
	go clean $(GOFLAGS) -i ./...

	cd $(BUILD_WEBAPP_DIR) && $(MAKE) clean

	rm -rf api/data
	rm -rf logs

	rm -rf Godeps/_workspace/pkg/

	rm -f **/zboxnow.log
	rm -f .prepare-go

nuke: clean clean-docker
	@echo BOOM

	rm -rf data

setup-mac:
	echo $$(boot2docker ip 2> /dev/null) dockerhost | sudo tee -a /etc/hosts

zbox: check-style build build-client
	@ echo Packaging ZBox Now!

	# Remove any old files
	rm -Rf $(DIST_ROOT)

	# Create needed directories
	mkdir -p $(DIST_PATH)/bin
	mkdir -p $(DIST_PATH)/logs

	# Copy binary
	@if [ "$(GOOS)" = "linux" ]; then \
		echo Packaging ZBox Now! for linux; \
		cp $(GOPATH)/bin/linux_amd64/platform $(DIST_PATH)/bin; \
	else \
		cp $(GOPATH)/bin/platform $(DIST_PATH)/bin; \
	fi

	# Resource directories
	cp -RL config $(DIST_PATH)
	cp -RL fonts $(DIST_PATH)
	cp -RL templates $(DIST_PATH)
	cp -RL i18n $(DIST_PATH)

	# Package webapp
	mkdir -p $(DIST_PATH)/webapp/dist
	cp -RL $(BUILD_WEBAPP_DIR)/dist $(DIST_PATH)/webapp
	mv $(DIST_PATH)/webapp/dist/bundle.js $(DIST_PATH)/webapp/dist/bundle-$(BUILD_NUMBER).js
	sed -i'.bak' 's|bundle.js|bundle-$(BUILD_NUMBER).js|g' $(DIST_PATH)/webapp/dist/root.html
	rm $(DIST_PATH)/webapp/dist/root.html.bak

	# Help files
ifeq ($(BUILD_ENTERPRISE_READY),true)
	cp $(BUILD_ENTERPRISE_DIR)/ENTERPRISE-EDITION-LICENSE.txt $(DIST_PATH)
else
	cp build/MIT-COMPILED-LICENSE.md $(DIST_PATH)
endif
	cp NOTICE.txt $(DIST_PATH)
	cp README.md $(DIST_PATH)

	@echo Done building ZBox NOW!

zbox-run:
	docker run --name ${DOCKER_CONTAINER_NAME} -d --publish 8065:80 --link=zbox-mysql:mysql --link=zboxOAuth2:dockerhost ${DOCKERNAME}:dev

zbox-dev:
	@echo Building DEV Docker Image
	@docker build -t ${DOCKERNAME}:dev -f docker/dev/Dockerfile .
	@rm -rf $(DIST_PATH)

zbox-prod:
	@echo Building PRODUCTION Docker Image
	@docker build -t ${DOCKERNAME}:${ZBOXNOW_VERSION} -f docker/prod/Dockerfile .
	@rm -rf $(DIST_PATH)
	@docker tag -f ${DOCKERNAME}:${ZBOXNOW_VERSION} ${DOCKERNAME}:latest; \

deploy:
	@echo "Pushing docker to registry"
	@docker push ${DOCKERNAME}
	@sudo rm -rf $(DIST_PATH)