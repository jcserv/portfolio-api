run/docker:
	docker compose up -d

run/local:
	go build ./cmd/go-api-template/main.go && ./main

clean:
	rm main
