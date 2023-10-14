import logging

import kserve
import base64
import io
import numpy as np
import pandas as pd

from typing import Dict
from PIL import Image
from ray import serve
from mmocr.utils.ocr import MMOCR


def convert_to_jpg(pil_image):
  open_cv_image = np.array(pil_image)
  open_cv_image = open_cv_image[:, :, ::-1].copy()
  image = open_cv_image

  if len(image.shape) > 2 and image.shape[2] == 4:
    # convert the image from RGBA2RGB
    image = cv2.cvtColor(image, cv2.COLOR_BGRA2BGR)
  return image


@serve.deployment(name="ocr", num_replicas=1)
class OCRModel(kserve.Model):
  def __init__(self):
    self.name = "ocr"
    super().__init__("ocr")
    self.load()

  def load(self):
    uri_config_dir = "/mmocr/configs"
    det_model = "MaskRCNN_IC17"
    reg_model = "ABINet"
    device = "cpu"

    ocr_model = MMOCR(
      det=det_model,
      recog=reg_model,
      device=device,
      config_dir=uri_config_dir
    )

    self.model = ocr_model
    self.ready = True

  def predict(self, payload: Dict) -> Dict:
    inputs = payload["instances"]
    data = inputs[0]["image"]["b64"]

    raw_img_data = base64.b64decode(data)
    input_image = Image.open(io.BytesIO(raw_img_data))
    image = convert_to_jpg(input_image)
    try:
      result = self.model.readtext(image, details=True,
                                   print_result=False, imshow=False)

      df = pd.DataFrame(result[0]['result'])
      df_sorted = df.sort_values(by=["box"])

      return {"predictions": df_sorted['text'].tolist()}
    except Exception:
      logging.debug(f"Error: {Exception}")
      return {"predictions": "Internal Error"}


if __name__ == "__main__":
  kserve.ModelServer().start({"ocr": OCRModel})
