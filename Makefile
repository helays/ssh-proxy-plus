.PHONY:  frontend build-arm64 build-arm64-dev build-amd64 build-amd64-dev
$(info start build)
GOARCH=$(shell go env GOARCH)
$(info cpu arch ${GOARCH})


srcFile= ./cmd/main.go
outFile= proxy-plus
execOut = ./runtime/build

TAG_PARTS ?= 2  # 默认拼接前两部分，可以通过 make TAG_PARTS=3 来覆盖

ifeq ($(OS),Windows_NT)
	IS_WINDOWS := 1
	BuildTime=$(shell echo %date% %time%)
    $(shell if not exist ${execOut} mkdir ${execOut})
    Version = $(shell powershell -NoProfile -Command "$$result = git describe --tags --always; $$parts = $$result.Split('-'); if ($$parts.Length -ge $(TAG_PARTS)) { [string]::Join('-', $$parts[0..($(TAG_PARTS)-1)]) } else { $$result }")
else
	BuildTime=$(shell date +"%Y-%m-%d %H:%M:%S")
    $(shell mkdir -p $(execOut) 2>/dev/null || true)
    Version=$(shell git describe --tags --always | awk -v parts=$(TAG_PARTS) 'BEGIN{FS="-"; OFS="-"} {if (NF >= parts) {for(i=1;i<=parts;i++) printf("%s%s", $$i, (i==parts?"":"-")); print ""} else {print $$0}}')
endif

$(info build version ${Version})

$(info build time ${BuildTime})

define LDFLAGS
"-X 'helay.net/go/utils/v3.Version=${Version}' \
-X 'helay.net/go/utils/v3.BuildTime=${BuildTime}' \
-linkmode external \
-extldflags=-static"
endef

$(info build params ${LDFLAGS})

BUILD_CMD = $(if $(IS_WINDOWS), \
    SET CGO_ENABLED=auto&SET GOOS=$(1)&SET GOARCH=$(2)&go build -tags "timetzdata $(if $(4),$(4))" -ldflags ${LDFLAGS} -o "${execOut}/${outFile}-${2}$(if $(findstring dev,$(4)),,-static)${lic_suffix}-${Version}" $(3), \
    CGO_ENABLED=auto CC=x86_64-linux-musl-gcc GOOS=$(1) GOARCH=$(2) go build -tags "timetzdata $(if $(4),$(4))" -ldflags ${LDFLAGS} -o "${execOut}/${outFile}-${2}$(if $(findstring dev,$(4)),,-static)${lic_suffix}-${Version}" $(3))

$(info prepare build)

# 编译部分
frontend:
	@echo "Building frontend..."
	@cd frontend && npm run build && cd ..
	@echo "Frontend build completed."
build-arm:
	$(call BUILD_CMD,linux,arm,${srcFile})
build-arm64:
	$(call BUILD_CMD,linux,arm64,${srcFile})
build-arm64-dev:
	$(call BUILD_CMD,linux,arm64,${srcFile},dev)
build-amd64:
	$(call BUILD_CMD,linux,$(GOARCH),${srcFile})
build-amd64-dev:
	$(call BUILD_CMD,linux,$(GOARCH),${srcFile},dev)
build: frontend  build-amd64
env:
	go env
#windows:
#	$(call BUILD_CMD,windows,$(GOARCH),shlt.data.handle)