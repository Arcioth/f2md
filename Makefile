PREFIX ?= $(HOME)/.local
BINDIR ?= $(PREFIX)/bin

.PHONY: all build install clean

all: build

build:
	go build -ldflags="-s -w" -o f2md .

install: build
	install -d $(DESTDIR)$(BINDIR)
	install -m 755 f2md $(DESTDIR)$(BINDIR)/f2md
	ln -sf f2md $(DESTDIR)$(BINDIR)/folder2md

clean:
	rm -f f2md folder2md
