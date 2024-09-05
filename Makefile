.PHONY: cover clean

cover:
	go test -v ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o /tmp/goglcover.html
	google-chrome /tmp/goglcover.html

clean:
	git clean -fx
