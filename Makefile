.POSIX:

ifndef PREFIX
  PREFIX = /usr/local
endif

clean:
	@rm -f ticktock

install:
	mkdir -p $(DESTDIR)$(PREFIX)/bin
	go build -o $(DESTDIR)$(PREFIX)/bin/

uninstall:
	rm -f $(DESTDIR)$(PREFIX)/bin/gem2site

.PHONY: install uninstall clean
