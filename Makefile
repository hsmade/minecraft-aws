.PHONY: build
build: storage_list

# FIXME: detect changes
.PHONY: storage_list
storage_list:
	docker build --platform x86_64 --build-arg app=storage_list -t $(ECR)/storage_list .
	docker push $(ECR)/storage_list

.PHONY: infrastructure
infrastructure:
	source .env && \
	cd infrastructure && \
    terraform init \
	  -backend-config="bucket=$(TFSTATE_BUCKET)" \
	  -backend-config="region=$(TFSTATE_REGION)" && \
	terraform apply