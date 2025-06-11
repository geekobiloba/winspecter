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
GOFILES_CLI := main.go collector.go stringer.go table.go cli.go text.go marshaler.go
GOFILES_GUI := main.go collector.go stringer.go table.go gui.go html.go

.PHONY: all
all: $(BIN)

$(BIN_CLI): $(GOFILES_CLI) $(ICON) $(COFF)
	go mod tidy
	go vet ./...
	go build -tags=cli -o $(BIN_CLI) -ldflags "-s -w" --trimpath -buildvcs=false .

$(BIN_GUI): $(GOFILES_GUI) $(TMPL) $(CSS) $(JS) $(ICON) $(COFF)
	go mod tidy
	go vet ./...
	go build -tags=gui -o $(BIN_GUI) -ldflags "-s -w -H=windowsgui" --trimpath -buildvcs=false .

$(COFF): $(ICON)
	rsrc -ico $<

$(ICON): $(LOGO)
	magick -define icon:auto-resize=256,128,64,48,32,16 $< $@

.PHONY: build
build: vet
	go build -tags=cli -o $(BIN_CLI) -ldflags "-s -w" --trimpath -buildvcs=false .
	go build -tags=gui -o $(BIN_GUI) -ldflags "-s -w -H=windowsgui" --trimpath -buildvcs=false .

.PHONY: vet
vet: tidy
	go vet ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: fmt
fmt:
	go fmt ./...


ifeq ($(OS),Windows_NT)

.PHONY: clean
clean:
	powershell.exe -NoProfile -Command \
		"Remove-Item -Force -ErrorAction Ignore *.html"

.PHONY: realclean
realclean: clean
	powershell.exe -NoProfile -Command \
		"Remove-Item -Force -ErrorAction Ignore *.exe, *.syso"

else

.PHONY: clean
clean:
	rm -f *.html

.PHONY: realclean
realclean: clean
	rm -f *.exe *.syso

endif
