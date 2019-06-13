cnf ?= config.env
include $(cnf)
export $(shell sed 's/=.*//' $(cnf))

dpl ?= deploy.env
include $(dpl)
export $(shell sed 's/=.*//' $(dpl))

DOCKER_REPO := $(AWS_ACCOUNT).dkr.ecr.$(AWS_REGION).amazonaws.com

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "DEV")

.DEFAULT_GOAL := build

test: build-dev run-test

run-test:
	docker run --rm -it -v $(PWD):/go/src/app --network dev --entrypoint "/usr/local/go/bin/go" $(APP_NAME)-dev test


dev: build-dev run-dev

forge-build:
	forge build

forge-deploy:
	forge deploy

build-dev:
	docker build -t $(APP_NAME)-dev -f docker/Dockerfile-dev .

run-dev:
	docker run --rm -it -p $(HOST_PORT):3000 -v $(PWD):/go/src/app --network dev $(APP_NAME)-dev

build:
	docker build -t ${APP_NAME} -f docker/Dockerfile .

build-nc:
	docker build --no-cache -t ${APP_NAME} -f docker/Dockerfile .

run:
	docker run -it --rm --env-file=./config.env -p=$(HOST_PORT):$(PORT) \
		--name="$(APP_NAME)" $(APP_NAME)

up: build run

stop:
	docker stop $(APP_NAME); docker rm $(APP_NAME)

release: test build-nc publish

publish: repo-login publish-latest publish-version

publish-latest: tag-latest
	@echo 'publish latest to $(DOCKER_REPO)'
	docker push $(DOCKER_REPO)/$(APP_NAME):latest

publish-version: tag-version
	@echo 'publish $(VERSION) to $(DOCKER_REPO)'
docker push $(DOCKER_REPO)/$(APP_NAME):$(VERSION)

tag: tag-latest tag-version

tag-latest:
	@echo 'create tag latest'
	docker tag $(APP_NAME) $(DOCKER_REPO)/$(APP_NAME):latest

tag-version:
	@echo 'create tag $(VERSION)'
	docker tag $(APP_NAME) $(DOCKER_REPO)/$(APP_NAME):$(VERSION)

CMD_REPOLOGIN := "eval $$\( aws ecr"
ifdef AWS_CLI_PROFILE
CMD_REPOLOGIN += " --profile $(AWS_CLI_PROFILE)"
endif
ifdef AWS_CLI_REGION
CMD_REPOLOGIN += " --region $(AWS_CLI_REGION)"
endif
CMD_REPOLOGIN += " get-login --no-include-email \)"

repo-login:
	@eval $(CMD_REPOLOGIN)

version:
	@echo $(VERSION)
