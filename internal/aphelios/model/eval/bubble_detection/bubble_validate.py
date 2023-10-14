import os

import click
import cv2
import numpy as np
import torch

from internal.aphelios.model.omr.python.model.bubble_detection.kserve.utils import get_model_object_serve, get_transform

MEAN = [0.485, 0.456, 0.406]
STD = [0.229, 0.224, 0.225]


@click.command()
@click.option('--image_test', prompt='folder image test set',
              help='folder image which be use for validation')
@click.option('--model_path', prompt='model weighted path', help='Model paths')
def validation(image_test, model_path):
  # 1. Load model
  # Set the device to be used for evaluation
  device = "cuda" if torch.cuda.is_available() else "cpu"
  num_classes = 3

  model = get_model_object_serve(num_classes, 0)
  checkpoint = torch.load(model_path, map_location=device)
  model.load_state_dict(checkpoint)
  model.to(device).eval()

  # 2. predict
  transforms = get_transform(False)
  for img_name in os.listdir(image_test):
    image = cv2.imread(f"{image_test}/{img_name}")
    image = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
    image = cv2.cvtColor(image, cv2.COLOR_GRAY2RGB)

    image_tensor = torch.tensor(image).permute(2, 0, 1)
    image_ = torch.from_numpy(np.array([image_tensor.numpy()])).to(device)
    image_nor = transforms(image_).to(device)

    output = model(image_nor)

    # 3. Print out format
    threshold = 0.8
    boxes = output[0]["boxes"].tolist()
    labels = output[0]["labels"].tolist()
    score = output[0]["scores"].tolist()

    file_name = img_name.split(".")[0]
    os.makedirs("predict", exist_ok=True)
    with open(f"predict/{file_name}.txt", "a+") as f:
      for i in range(len(boxes)):
        if score[i] > threshold:
          line = ""
          # 1. label
          line = line + f"{labels[i]}" + " "
          # 2. score
          line = line + f"{score[i]}" + " "
          # 3. x1, y1, x2, y2
          line = line + f"{boxes[i][0]}" + " " + f"{boxes[i][1]}" + " " + f"{boxes[i][2]}" + " " + f"{boxes[i][3]}"

          f.writelines(line + '\n')
          print(f"{file_name}.txt - {line}")


if __name__ == '__main__':
  validation()
