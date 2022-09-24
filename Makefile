.PHONY: build
build: storage_list

# FIXME: detect changes
.PHONY: storage_list
storage_list:
	docker build --platform linux/amd64 --build-arg app=storage_list -t $(ECR)/storage_list .
	docker push $(ECR)/storage_list

.PHONY: scripts
scripts:
	docker build --platform linux/amd64 -t $(ECR)/scripts -f scripts.Dockerfile .
	docker push $(ECR)/scripts

.PHONY: infrastructure
infrastructure:
	cd infrastructure && \
    terraform init \
	  -backend-config="bucket=$(TFSTATE_BUCKET)" \
	  -backend-config="region=$(TFSTATE_REGION)" && \
	terraform apply -auto-approve -input=false