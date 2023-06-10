###########################
# add your config here
###########################
export SPOTIFY_CLIENT_ID =
export SPOTIFY_PLAYLIST_ID =
export SPOTIFY_CLIENT_SECRET =
export SPOTIFY_REFRESH_TOKEN =
export PLAYLIST_SIZE = 30
###########################
# static config
###########################
CODE_FOLDERS=cmd internal pkg
.PHONY: bot bot-container fmt imports lint test

bot:
	go run cmd/main.go

bot-container:
	docker build .
	docker run

fmt: ## Run gofumpt on all the source code
	@echo "${COLOUR_GREEN}run following command to ensure you have the latest gofumpt installed${COLOUR_NORMAL}"
	@echo "go install mvdan.cc/gofumpt@latest"
	gofumpt -l -w $(CODE_FOLDERS)

lint: imports fmt ## Runs the golangci-lint checker
	@echo "run following command to ensure you have the latest golangci-lint installed."
	@echo "current version"
	@echo "go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
	golangci-lint run -v

test:
	go test ./...