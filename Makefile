VERSION := 0.0.25
FNAME := uv4-middleware
GNAME := awculle_datahub
KO_DOCKER_REPO := strunck

templ:
	templ generate

run:
	KO_DATA_PATH=kodata \
	go run -ldflags="-X 'main.Version=${VERSION}'" \
	main.go

dev: 
	@zsh ./devContainer.sh $(VERSION)

debug-nats:
	podman run --pod uv4pod --rm -it natsio/nats-box:latest /bin/sh
	
ko: systemctl
	DOCKER_HOST=unix:///run/user/1000/podman/podman.sock \
	KO_DOCKER_REPO=${KO_DOCKER_REPO} \
	DATA_PATH=kodata \
	GOFLAGS="-ldflags=-X=main.Version=${VERSION}" \
	ko build --local -B 

koarm: systemctl
	DOCKER_HOST=unix:///run/user/1000/podman/podman.sock \
	KO_DOCKER_REPO=${KO_DOCKER_REPO} \
	DATA_PATH=kodata \
	GOFLAGS="-ldflags=-X=main.Version=${VERSION}" \
	ko build --local -B --platform=linux/arm64 --tags=arm64 

login:
	podman login docker.io	

systemctl:
	systemctl --user start podman.service

release: koarm ko
	buildah manifest create ${FNAME}:$(VERSION)
	buildah manifest add ${FNAME}:$(VERSION) ${KO_DOCKER_REPO}/${FNAME}:arm64
	buildah manifest add ${FNAME}:$(VERSION) ${KO_DOCKER_REPO}/${FNAME}:latest
	buildah manifest push --all --rm ${FNAME}:$(VERSION) docker://docker.io/strunck/${FNAME}:$(VERSION)

vet:
	go vet ./...

all: templ vet run
