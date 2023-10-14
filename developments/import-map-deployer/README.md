# import-map-deployers

eval $(minikube docker-env)
DOCKER_BUILDKIT=1 docker build --progress=plain \
 --tag asia.gcr.io/student-coach-e1e95/import-map-deployer:0.0.2 \
 -f import-map-deployer.Dockerfile .

docker push asia.gcr.io/student-coach-e1e95/import-map-deployer:0.0.2

docker run -p 5000:5000 asia.gcr.io/student-coach-e1e95/import-map-deployer:0.0.2

# Example script

http --verify=no -a admin:M@nabie123 https://admin.local-green.manabie.io:31600/imd/environments

http --verify=no -a admin:M@nabie123 https://admin.local-green.manabie.io:31600/imd/import-map.json?env=mamabie

http --verify=no -a admin:M@nabie123 PATCH https://admin.staging.manabie.io/imd/services?env=manabie service=react-dom url="https://cdn.jsdelivr.net/npm/@esm-bundle/react-dom@17.0.2-fix.0/esm/react-dom.development.min.js"


http --verify=no -a admin:M@nabie123 PATCH https://admin.staging.manabie.io/imd/services?env=manabie service=react url="https://cdn.jsdelivr.net/npm/@esm-bundle/react@17.0.2-fix.0/esm/react.development.min.js"
