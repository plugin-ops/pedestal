override PROJECT_NAME 			= pedestal


swagger:
	GOARCH=amd64 go build -o ${shell pwd}/bin/swag ${shell pwd}/build/swag/main.go
	rm -rf ./$(PROJECT_NAME)/docs
	${shell pwd}/bin/swag init -g ./$(PROJECT_NAME)/serve/http/app.go -o ./$(PROJECT_NAME)/docs
