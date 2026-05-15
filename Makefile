VERSION := 0.0.26
FNAME := uv4-middleware
KO_DOCKER_REPO := strunck
REGISTRY_NAME ?= local-registry
REGISTRY_IMAGE ?= docker.io/library/registry:2
REGISTRY_PORT ?= 5000
LOCAL_REGISTRY ?= localhost:${REGISTRY_PORT}
LOCAL_IMAGE := ${LOCAL_REGISTRY}/${FNAME}
DOCKER_REGISTRY := docker.io/${KO_DOCKER_REPO}/${FNAME}

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

ensure-registry:
	@if curl -fsS http://${LOCAL_REGISTRY}/v2/ >/dev/null 2>&1; then \
		echo "Local registry ${LOCAL_REGISTRY} is already running."; \
	elif podman container exists ${REGISTRY_NAME}; then \
		echo "Starting existing local registry ${REGISTRY_NAME} on ${LOCAL_REGISTRY}..."; \
		podman start ${REGISTRY_NAME}; \
	else \
		echo "Creating local registry ${REGISTRY_NAME} on ${LOCAL_REGISTRY}..."; \
		podman run -d --name ${REGISTRY_NAME} -p ${REGISTRY_PORT}:5000 --restart=always ${REGISTRY_IMAGE}; \
	fi
	@curl -fsS http://${LOCAL_REGISTRY}/v2/ >/dev/null

multi: ensure-registry
	podman build --platform linux/amd64 -f Dockerfile.prod --build-arg VERSION=${VERSION} -t ${LOCAL_IMAGE}:${VERSION}-amd64 .
	podman build --platform linux/arm64 -f Dockerfile.prod --build-arg VERSION=${VERSION} -t ${LOCAL_IMAGE}:${VERSION}-arm64 .
	buildah manifest rm ${LOCAL_IMAGE}:${VERSION} || true
	buildah manifest create ${LOCAL_IMAGE}:${VERSION}
	buildah manifest add ${LOCAL_IMAGE}:${VERSION} containers-storage:${LOCAL_IMAGE}:${VERSION}-amd64 --tls-verify=false
	buildah manifest add ${LOCAL_IMAGE}:${VERSION} containers-storage:${LOCAL_IMAGE}:${VERSION}-arm64 --tls-verify=false
	buildah manifest push --all ${LOCAL_IMAGE}:${VERSION} docker://${LOCAL_IMAGE}:${VERSION} --tls-verify=false

pushdo: login
	buildah manifest push --rm --all ${LOCAL_IMAGE}:${VERSION} docker://${DOCKER_REGISTRY}:${VERSION}
			

mani:
	buildah manifest exists ${LOCAL_IMAGE}:${VERSION}

	
ko: systemctl
	DOCKER_HOST=unix:///run/user/1000/podman/podman.sock \
	KO_DOCKER_REPO=${KO_DOCKER_REPO}/${FNAME} \
	DATA_PATH=kodata \
	GOFLAGS="-ldflags=-X=main.Version=${VERSION}" \
	ko build --platform=linux/amd64,linux/arm64 \
	--local main.go

koarm: systemctl
	DOCKER_HOST=unix:///run/user/1000/podman/podman.sock \
	KO_DOCKER_REPO=${KO_DOCKER_REPO} \
	DATA_PATH=kodata \
	GOFLAGS="-ldflags=-X=main.Version=${VERSION}" \
	ko build --local -B --platform=linux/arm64 --tags=${VERSION} \
	main.go

login:
	podman login docker.io	

systemctl:
	systemctl --user start podman.service

release: koarm ko		
	buildah manifest rm ${LOCAL_IMAGE}:${VERSION} || true
	buildah manifest create ${LOCAL_IMAGE}:${VERSION}
	buildah manifest add ${LOCAL_IMAGE}:${VERSION} ${KO_DOCKER_REPO}/${FNAME}:${VERSION}-arm64 --tls-skip-verify
	buildah manifest add ${LOCAL_IMAGE}:${VERSION} ${KO_DOCKER_REPO}/${FNAME}:${VERSION}-amd64 --tls-skip-verify
	buildah manifest push --all --rm ${LOCAL_IMAGE}:${VERSION} docker://docker.io/strunck/${FNAME}:$(VERSION)

vet:
	go vet ./...

all: templ vet run
