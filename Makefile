ifneq (,$(wildcard ./.env))
	include .env
	export
else
  $(error You haven't setup your .env file. Please refer to the readme)
endif
LAMBDA_GOOS=linux
LAMBDA_GOARCH=arm64
LAMBDA_GOCC?=go
LAMBDA_GOFLAGS=-tags=lambda.norpc
LAMBDA_CGO_ENABLED=0
LAMBDAS=build/aggregatesubmitter/bootstrap build/getblob/bootstrap build/getclaim/bootstrap build/getroot/bootstrap build/pieceaccepter/bootstrap build/pieceaggregator/bootstrap build/postroot/bootstrap build/putblob/bootstrap

.PHONY: clean-lambda

clean-lambda:
	rm -rf build

.PHONY: clean-terraform

clean-terraform:
	tofu -chdir=deploy/app destroy

.PHONY: clean

clean: clean-terraform clean-lambda

lambdas: $(LAMBDAS)

.PHONY: $(LAMBDAS)

$(LAMBDAS): build/%/bootstrap:
	GOOS=$(LAMBDA_GOOS) GOARCH=$(LAMBDA_GOARCH) CGO_ENABLED=$(LAMBDA_CGO_ENABLED) $(LAMBDA_GOCC) build $(LAMBDA_GOFLAGS) -o $@ cmd/lambda/$*/main.go

deploy/app/.terraform:
	tofu -chdir=deploy/app init

.tfworkspace: deploy/app/.terraform
	tofu -chdir=deploy/app workspace new $(TF_WORKSPACE)
	touch .tfworkspace

.PHONY: init

init: deploy/app/.terraform .tfworkspace

.PHONY: validate

validate: deploy/app/.terraform .tfworkspace
	tofu -chdir=deploy/app validate

.PHONY: plan

plan: deploy/app/.terraform .tfworkspace $(LAMBDAS)
	tofu -chdir=deploy/app plan

.PHONY: apply

apply: deploy/app/.terraform .tfworkspace $(LAMBDAS)
	tofu -chdir=deploy/app apply


deploy/app/.terraform:
	tofu -chdir=deploy/app init

shared:
