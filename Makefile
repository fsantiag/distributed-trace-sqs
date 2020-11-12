build:
	cd producer && GOOS=linux GOARCH=amd64 go build -o producer-app
	cd consumer && GOOS=linux GOARCH=amd64 go build -o consumer-app
