run/docker:
	docker compose up -d

run/local:
	go build ./cmd/portfolio-api/main.go && ./main

# Export all variables from .env file to the current shell
# Usage: source <(make exportenv)
exportenv:
	@if [ -f .env ]; then \
		echo "# Exporting environment variables from .env file..."; \
		while IFS='=' read -r key value || [ -n "$$key" ]; do \
			if [ ! -z "$$key" ] && ! echo "$$key" | grep -q '^#'; then \
				value=$$(echo "$$value" | sed -e 's/^"//' -e 's/"$$//' -e "s/^'//" -e "s/'$$//"); \
				echo "export $$key=$$value"; \
			fi; \
		done < .env; \
		echo "# Environment variables exported successfully."; \
	else \
		echo "# Error: .env file not found in current directory" >&2; \
		exit 1; \
	fi

clean:
	rm main && rm ./internal/db/portfolio-api.db