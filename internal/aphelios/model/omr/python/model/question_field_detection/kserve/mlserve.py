import io

import kserve
import torch
import base64
import mlflow

import numpy as np

from typing import Dict
from PIL import Image
from ray import serve
from mmcv import Config
from mmdet.apis import inference_detector

def convert_to_jpg(pil_image):
  open_cv_image = np.array(pil_image)
  open_cv_image = open_cv_image[:, :, ::-1].copy()
  image = open_cv_image

  return image

@serve.deployment(name="question-field", num_replicas=1)
class QuestionFieldModel(kserve.Model):
  def __init__(self):
    self.name = "question-field"
    super().__init__("question-field")
    self.load()

  def load(self):
    uri_config = "./mm_configs.py"
    uri_model = "/mnt/models"

    device = "cpu"

    model = mlflow.pytorch.load_model(uri_model, map_location=torch.device(device))
    model.eval()

    config = Config.fromfile(uri_config)
    model.module.cfg = config

    self.model = model
    self.ready = True

  def predict(self, payload: Dict) -> Dict:
    inputs = payload["instances"]
    data = inputs[0]["image"]["b64"]

    raw_img_data = base64.b64decode(data)
    input_image = Image.open(io.BytesIO(raw_img_data))
    image = convert_to_jpg(input_image)
    try:
      result = inference_detector(self.model.module, image)
      return {"predictions": result[0].tolist()}
    except:
      return {"predictions": "Internal Error"}



if __name__ == "__main__":
  kserve.ModelServer().start({"question-field": QuestionFieldModel})