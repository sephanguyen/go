# graphql-mesh

eval $(minikube docker-env)
DOCKER_BUILDKIT=1 docker build --progress=plain \
 --tag asia.gcr.io/student-coach-e1e95/graphql-mesh:0.0.1 \
 -f graphql-mesh.Dockerfile .

docker push asia.gcr.io/student-coach-e1e95/graphql-mesh:0.0.1

docker run -p 5000:5000 asia.gcr.io/student-coach-e1e95/graphql-mesh:0.0.1
