NAME=opentelemetry-lambda-extension-with-elasticexporter
REGIONS=eu-west-1
GIT_HASH=$(shell git rev-parse --short HEAD)
GO_EXEC=go

build:
	GOOS=linux GOARCH=amd64 ${GO_EXEC} build -o bin/extensions/$(NAME) *.go
	chmod +x bin/extensions/$(NAME)

publish:
	rm -f bin/*.zip
	cd bin && zip -r $(GIT_HASH).zip extensions/
	for region in $(REGIONS); do \
		aws --region $$region s3 cp ./bin/$(GIT_HASH).zip s3://$(S3_BUCKET)-$$region/$(NAME)/; \
		aws lambda publish-layer-version \
			--layer-name "$(NAME)" \
			--description "OpenTelemetry with otlp receiver and elastic exporter" \
			--region $$region \
			--content S3Bucket=$(S3_BUCKET)-$$region,S3Key=$(NAME)/$(GIT_HASH).zip \
			--compatible-runtimes nodejs10.x nodejs12.x python3.6 python3.7 python3.8 ruby2.5 ruby2.7 java8 java8.al2 java11 dotnetcore3.1 provided.al2 \
			--no-cli-pager \
			--output text ; \
	done

public:
	for region in $(REGIONS); do \
		aws lambda add-layer-version-permission \
			--layer-name $(NAME)  \
			--principal '*'  \
			--action lambda:GetLayerVersion \
			--version-number $(VERSION) \
			--statement-id public \
			--region $$region \
			--no-cli-pager \
			--output text ; \
	done

build-OpenTelemetryExtensionLayer:
	GOOS=linux GOARCH=amd64 go build -o $(ARTIFACTS_DIR)/extensions/$(NAME) *.go
	chmod +x $(ARTIFACTS_DIR)/extensions/$(NAME)

run-OpenTelemetryExtensionLayer:
	go run $(NAME)/*.go

test:
	# TODO
