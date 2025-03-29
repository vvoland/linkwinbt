TAG?=$(shell git describe --tags --exact-match 2>/dev/null || git describe --always)

.PHONY: run
run: img
	docker run --rm -it \
		-v /mnt:/mnt:ro \
		-v /var/lib/bluetooth:/var/lib/bluetooth \
		vlnd/linkwinbt:$(TAG) /mnt

.PHONY: img
img:
	docker build . -t vlnd/linkwinbt:$(TAG) -t vlnd/linkwinbt:latest

.PHONY: publish
publish: img
	docker push vlnd/linkwinbt:$(TAG)
	docker push vlnd/linkwinbt:latest
	
