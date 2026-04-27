build:
	@if [ "$(OS)" = "Windows_NT" ]; then \
		go build -o vessel.exe; \
	else \
		go build -o vessel; \
	fi

install: build
	sudo mv vessel /usr/local/bin/vessel