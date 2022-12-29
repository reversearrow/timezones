build:
	docker build . -t timezones

run:
	docker run -p 8080:8080 timezones