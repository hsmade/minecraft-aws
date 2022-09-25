LAMBDA_HANDLERS = server_start server_stop server_status servers_list

$(LAMBDA_HANDLERS):
	docker build --platform linux/amd64 --build-arg app=$% -t $(ECR)/$% .
	docker push $(ECR)/$%

.PHONY: infrastructure
infrastructure:
	cd infrastructure && \
    terraform init \
	  -backend-config="bucket=$(TFSTATE_BUCKET)" \
	  -backend-config="region=$(TFSTATE_REGION)" && \
	terraform apply -auto-approve -input=false