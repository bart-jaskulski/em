build:
	go build -o ./bin/em main.go

install:
	cp ./bin/em /usr/local/bin/em
	cp ./bin/em-tmux /usr/local/bin/em-tmux

uninstall:
	rm /usr/local/bin/em
	rm /usr/local/bin/em-tmux

.PHONY: build install

default: build

