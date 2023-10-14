
# Instruction (v0.9.0).
This deployment support for deploy model ml from mlflow on k8s with serverless. After training model and register with mlflow, kserve will pull image from mlflow 's artifact to serving both HTTP(port 8080) and GPRC (port 9000)
For deploy model, we configure yaml file in folder `./model/modelconfigure`

# To test example: 
Watiting until the service pods is ready, then we can forward port to test or you ingress dns for serving. 
- Test endpoint:  
  - for forward port to 8080: 
    ``` 
    # locate at: `backend/deployments/helm/platforms/machinelearning/kserve`
    curl -v \            
    -H "Content-Type: application/json" \
    -d @./model/modelconfigure/manabie/local/example_mana_input.json \
    http://localhost:8080/v2/models/ml-manabie/infer 
    ```

More detail: [Official KServe Doc](https://kserve.github.io/website/0.9/get_started/first_isvc/#5-perform-inference)