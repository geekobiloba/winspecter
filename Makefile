################################################################################
# Winspecter Makefile
################################################################################

BIN_GUI := winspecter.exe
BIN_CLI := winspecter-cli.exe
BIN     := $(BIN_CLI) $(BIN_GUI)
TMPL := assets/html.tmpl
CSS  := assets/style.css
JS   := assets/script.js
COFF := rsrc_windows_amd64.syso
ICON := assets/favicon.ico
LOGO := assets/winspecter.png
GOFILES_CLI := main.go collector.go string.go table.go cli.go text.go serial.go
GOFILES_GUI := main.go collector.go string.go table.go gui.go html.go

.PHONY: all cli gui build vet tidy fmt clean realclean clean realclean

all: $(BIN)

cli: $(BIN_CLI)

$(BIN_CLI): $(GOFILES_CLI) $(COFF)
	go mod tidy
	go vet ./...
	go build -tags=cli -o $(BIN_CLI) -ldflags "-s -w" --trimpath -buildvcs=false .
	upx --force-overwrite --best --lzma $@

gui: $(BIN_GUI)

$(BIN_GUI): $(GOFILES_GUI) $(TMPL) $(CSS) $(JS) $(ICON) $(COFF)
	go mod tidy
	go vet ./...
	go build -tags=gui -o $(BIN_GUI) -ldflags "-s -w -H=windowsgui" --trimpath -buildvcs=false .
	upx --force-overwrite --best --lzma $@

$(COFF): $(ICON)
	rsrc -ico $<

$(ICON): $(LOGO)
	magick -define icon:auto-resize=256,128,64,48,32,16 $< $@

build: vet
	go build -tags=cli -o $(BIN_CLI) -ldflags "-s -w" --trimpath -buildvcs=false .
	go build -tags=gui -o $(BIN_GUI) -ldflags "-s -w -H=windowsgui" --trimpath -buildvcs=false .
	upx --force-overwrite --best --lzma $(BIN_CLI)
	upx --force-overwrite --best --lzma $(BIN_GUI)

vet: tidy
	go vet ./...

tidy:
	go mod tidy

fmt:
	go fmt ./...

clean:
ifeq ($(OS),Windows_NT)
	powershell.exe -NoProfile -Command \
		"Remove-Item -Force -ErrorAction Ignore *.html"

else
	rm -f *.html
endif

realclean: clean
ifeq ($(OS),Windows_NT)
	powershell.exe -NoProfile -Command \
		"Remove-Item -Force -ErrorAction Ignore *.exe, *.syso"
else
	rm -f *.exe *.syso

endif
