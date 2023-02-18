LAMBDA_HANDLERS = servers_list server_start server_stop server_status
JS_FILE = infrastructure/site/js/app*js
$(LAMBDA_HANDLERS):
	docker build --platform linux/amd64 --build-arg app=$@ -t $(ECR)/$@ .
	docker push $(ECR)/$@

.PHONY: infrastructure
infrastructure:
	cd infrastructure && \
    terraform init \
	  -backend-config="bucket=$(TFSTATE_BUCKET)" \
	  -backend-config="region=$(TFSTATE_REGION)" && \
	terraform apply -auto-approve -input=false

infrastructure/site:
	cd web && \
	npm install && \
	npm run build && \
	mv dist ../infrastructure/site
	for file in infrastructure/site/js/app*js; do mv -v "$${file}" "$${file}.tmpl"; done
