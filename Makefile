TAG?=$(shell git describe --tags --exact-match 2>/dev/null || git describe --always)

.PHONY: run
run: img
	docker run --rm -it \
		-v /mnt:/mnt:ro \
		-v /var/lib/bluetooth:/var/lib/bluetooth \
		vlnd/linkwinbt:$(TAG) /mnt

.PHONY: img
img:
	docker build . -t vlnd/linkwinbt:$(TAG) -t vlnd/linkwinbt:latest \
		--attest type=provenance,mode=max --attest type=sbom \
		-o type=image,compression=zstd

.PHONY: publish
publish: img
	docker push vlnd/linkwinbt:$(TAG)
	docker push vlnd/linkwinbt:latest
	
