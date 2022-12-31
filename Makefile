.PHONY: build
build:
	./scripts/build-and-push.sh

.PHONY: run
run:
	docker run -p 8080:8080 timezones