# bufbuild

```sh
DOCKER_BUILDKIT=1 docker build --progress=plain \
 --build-arg GITHUB_TOKEN=${GITHUB_TOKEN} \
 --tag asia-docker.pkg.dev/student-coach-e1e95/manaverse/bufbuild:0.0.4 \
 -f bufbuild.Dockerfile .

docker push asia-docker.pkg.dev/student-coach-e1e95/manaverse/bufbuild:0.0.4
```