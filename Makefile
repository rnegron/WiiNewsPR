.PHONY: build run deploy clean

build:
	go build

run: build
	./WiiNewsPR

deploy:
	@echo "Building Lambda handler..."
	@cd deploy && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap handler.go

	@echo "Building WiiNewsPR binary..."
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o deploy/WiiNewsPR .

	@echo "Deploying to AWS..."
	@cd deploy && serverless deploy

	@echo "Deployed!"

remove:
	cd deploy && serverless remove

clean:
	rm -f WiiNewsPR
	rm -f deploy/bootstrap deploy/WiiNewsPR
