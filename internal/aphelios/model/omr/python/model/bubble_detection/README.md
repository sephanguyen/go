# 1.Instruction 
This folder for training and deploy model bubble (answer) detection.

# 2.How to use
In `/bubble_dectection`, there are 2 folder `bubble_dectection/kserve` and `bubble_dectection/training`. 
    - `bubble_detection/kserve` where we build service docker base on kserve server wraper.
    - `bubble_dectection/training` includes code for training/predict of model. The result of model when tracking will be store on mlflow server. 


# 3.Setup:

- Start mlflow server on staging env. 
  - Because mlflow server is not support yet for authentication, so we will port forwarding from k8s by kubectl.
    ```
    kubectl port-forward mlflow-59bd44cffd-7pt8q -n stag-manabie-machine-learning 5000:5000
    ```
- Install Environment:
  - Setup pytorch and CUDA on machine. torch.version.cuda 10.2
  - VRAM GPU at least 6GB. 
  - install essential lib in `training/requirement.txt`
- Data:
  - Format data is COCO format, include one folder image and one json result file. [COCO format](https://medium.com/mlearning-ai/coco-dataset-what-is-it-and-how-can-we-use-it-e34a5b0c6ecd#:~:text=What's%20the%20COCO%20format%3F,%2C%20bounding%20boxes%2C%20and%20bitmasks.)

# 4. Run.

- At `internal/aphelios/model/omr/python/model/bubble_detection/training` Run:
```
python maskRCNN.py
```
- Status of training phase will be sent to mlflow server at url: `http://localhost:5000` (please make sure you have forward port from kubectl) [Image](https://www.mlflow.org/docs/latest/_images/tutorial-compare.png)
  
- After training phase success, the checkpoint file will be saved at `artifacts` of this running. [Image](https://www.mlflow.org/docs/latest/_images/tutorial-artifact.png)

- For more detail, please check [mlflow official doc](https://www.mlflow.org/docs/latest/tutorials-and-examples/tutorial.html)