CONFIG_PATH=${HOME}/.storymetadatagenerator/
.PHONY: init
init:
		mkdir -p ${CONFIG_PATH}
.PHONY: gencert
gencert:
		make init
		cfssl gencert \
				-initca test/ca-csr.json | cfssljson -bare ca
		cfssl gencert \
				-ca=ca.pem \
				-ca-key=ca-key.pem \
				-config=test/ca-config.json \
				-profile=server \
				test/server-csr.json | cfssljson -bare server
		 cfssl gencert \
                -ca=ca.pem \
                -ca-key=ca-key.pem \
                -config=test/ca-config.json \
                -profile=client \
                -cn="nobody" \
                test/client-csr.json | cfssljson -bare nobody-client
		cfssl gencert \
        		-ca=ca.pem \
        		-ca-key=ca-key.pem \
        		-config=test/ca-config.json \
        		-profile=client \
        		-cn="root" \
        		test/client-csr.json | cfssljson -bare root-client
		mv *.pem *.csr ${CONFIG_PATH}
.PHONY: compile
compile:
		protoc api/v1/grpc/*.proto \
				--go_out=. \
				--go-grpc_out=. \
				--go_opt=paths=source_relative \
				--go-grpc_opt=paths=source_relative \
				--proto_path=.
$(CONFIG_PATH)/model.conf:
		cp test/model.conf ${CONFIG_PATH}/model.conf
$(CONFIG_PATH)/policy.csv:
		cp test/policy.csv ${CONFIG_PATH}/policy.csv
.PHONY: test
test: $(CONFIG_PATH)/model.conf $(CONFIG_PATH)/policy.csv
		@echo "building docker containers..."
		@make docker-start
		@echo "running tests..."
		go test -v -race ./...
		@echo "cleaning up docker containers..."
		@make docker-down
.PHONY: start
start:
	@make clean
	@make build
	@echo "running main program..."
	@./storymetadatagenerator --config-file=./config.json
.PHONY: docker-start
docker-start:
	@docker-compose --project-name=storymetadatagenerator up -d
.PHONY: docker-down
docker-down:
	@docker-compose --project-name=storymetadatagenerator down
.PHONY: build
build:
	@make test
	@go build ./cmd/storymetadatagenerator
.PHONY: clean
clean:
	@go clean -i
.PHONY: cover
cover:
	@mkdir .coverage || echo "hidden coverage folder exists"
	@echo "building docker containers..."
	@make docker-start
	@go test -v -cover ./... -coverprofile .coverage/coverage.out
	@go tool cover -html=.coverage/coverage.out -o .coverage/coverage.html
	@echo "cleaning up docker containers..."
	@make docker-down
.PHONY: covero
covero:
	@make cover
	@open .coverage/coverage.html
files += $(filter-out %_test.go,$(wildcard api/v1/storymetadata.go))
.PHONY: mocks
mocks:
	@rm -rf mocks
	@go install github.com/golang/mock/mockgen
	@echo "generating mocks..."
	@$(foreach file, $(files), mockgen -package mocks -destination mocks/$(subst /,'-',$(file)) -source $(file);)
