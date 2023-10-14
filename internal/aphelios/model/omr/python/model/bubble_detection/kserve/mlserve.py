import base64
import io
from typing import Dict

import cv2
import kserve
import numpy as np
import torch
from PIL import Image
from ray import serve
from utils import convert_to_jpg, get_model_object_serve, get_transform

MEAN = [0.485, 0.456, 0.406]
STD = [0.229, 0.224, 0.225]


@serve.deployment(name="bubble-model", num_replicas=1)
class BubbleDetectionModel(kserve.Model):
  def __init__(self):
    self.name = "bubble-model"
    super().__init__("bubble-model")
    self.load()

  def load(self):
    model_path = "/mnt/models/state_dict.pth"
    # Set the device to be used for evaluation
    device = 'cpu'
    num_classes = 3

    model = get_model_object_serve(num_classes, 0)
    checkpoint = torch.load(model_path, map_location=device)
    model.load_state_dict(checkpoint)
    model.to(device).eval()

    self.model = model
    self.ready = True

  def predict(self, payload: Dict) -> Dict:
    transforms = get_transform(False)
    device = 'cpu'
    inputs = payload["instances"]
    data = inputs[0]["image"]["b64"]

    raw_img_data = base64.b64decode(data)
    input_image = Image.open(io.BytesIO(raw_img_data))

    image = convert_to_jpg(input_image)
    image = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
    image = cv2.cvtColor(image, cv2.COLOR_GRAY2RGB)

    model = self.model

    image_tensor = torch.tensor(image).permute(2, 0, 1)

    image_ = torch.from_numpy(np.array([image_tensor.numpy()])).to(device)
    image_nor = transforms(image_).to(device)

    output = model(image_nor)

    print(f"output: {output}")
    return {"predictions": {
      "boxes": output[0]["boxes"].tolist(),
      "labels": output[0]["labels"].tolist(),
      "scores": output[0]["scores"].tolist()
    }}


if __name__ == "__main__":
  kserve.ModelServer().start({"bubble-model": BubbleDetectionModel})
