APP_EXECUTABLE?=./bin/catalog
RELEASE?=1.0
MIGRATIONS_RELEASE?=1.0
MIGRATIONS_IMAGENAME?=arahna/catalog-service-migrations:v$(MIGRATIONS_RELEASE)
IMAGENAME?=arahna/catalog-service:v$(RELEASE)

.PHONY: clean
clean:
	rm -f ${APP_EXECUTABLE}

.PHONY: build
build: clean
	docker build -t $(MIGRATIONS_IMAGENAME) -f DockerfileMigrations .
	docker build -t $(IMAGENAME) .

.PHONY: release
release:
	git tag v$(RELEASE)
	git push origin v$(RELEASE)