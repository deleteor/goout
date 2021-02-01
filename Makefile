APP := goout
GO := go

help:
	$(GO) run main.go -h

local:
	$(GO) build -v -o bundles/$(APP) main.go
