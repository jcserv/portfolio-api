run/docker:
	docker compose up -d

run/local:
	go build ./cmd/portfolio-api/main.go && ./main

clean:
	rm main
