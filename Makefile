APP := goout
GO := go

help:
	$(GO) run main.go -h

local:
	$(GO) build -v -o build/$(APP) main.go
