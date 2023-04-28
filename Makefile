deploy:
	cd server; GOOS=linux GOARCH=arm64 go build main.go; scp ./main ubuntu@54.164.224.210:main;