import json
import os
import shutil

import click

from internal.aphelios.data.augmentation.coco_format import cocoformatter_from_dict, cocoformatter_to_dict


def load(data_path):
  with open(data_path) as json_file:
    ann_dict = json.load(json_file)
    coco = cocoformatter_from_dict(ann_dict)
    print("DONE")
  return coco


def coco2map(images, bboxs, categories, image_ids, root_output):
  if os.path.isdir(root_output):
    shutil.rmtree(root_output)
  os.makedirs(root_output, exist_ok=True)
  for i in range(0, len(bboxs)):
    filename = images[image_ids[i]]
    filename = filename.split("/")[-1].split(".")[0]
    with open(f"{root_output}/{filename}.txt", "a") as file:
      line = ""

      # 0. A[0]: Category.
      line = line + f"{categories[i]}" + " "

      # 1. A[1:5]: x1 y1 x2 y2
      x1 = bboxs[i][0]
      y1 = bboxs[i][1]
      x2 = bboxs[i][0] + bboxs[i][2]
      y2 = bboxs[i][1] + bboxs[i][3]
      line = line + f"{x1} {y1} {x2} {y2}"

      # 3. A[9]: Difficulty.
      # line = line + "0"

      file.writelines(line + '\n')
      print(f"{filename}.txt - {line}")


def res2map(images, bboxs, categories, image_ids, root_output):
  if os.path.isdir(root_output):
    shutil.rmtree(root_output)
  os.makedirs(root_output, exist_ok=True)

  for i in range(0, len(bboxs)):
    filename = images[image_ids[i]]
    filename = filename.split("/")[-1].split(".")[0]
    with open(f"{root_output}/{filename}.txt", "a") as file:
      line = ""

      # 1. A[0]: Category.
      line = line + f"{categories[i]}" + " "

      # 2. A[1]: score.
      line = line + f"0.9" + " "

      # 3. A[2:6]: x1 y1 x2 y2
      x1 = bboxs[i][0]
      y1 = bboxs[i][1]
      x2 = bboxs[i][0] + bboxs[i][2]
      y2 = bboxs[i][1] + bboxs[i][3]
      line = line + f"{x1} {y1} {x2} {y2}"

      file.writelines(line + '\n')
      print(f"{filename}.txt - {line}")


@click.command()
@click.option('--ann_path', prompt='annotation file path',
              help='annotation file in coco format path (result.json)')
@click.option('--groundtruth_dir', prompt='groundtruth dir', help='Groundtruth dir in format.')
def eval(ann_path, groundtruth_dir):
  coco = load(ann_path)

  annotation = cocoformatter_to_dict(coco)["annotations"]
  images = cocoformatter_to_dict(coco)["images"]

  images_name = []
  bboxs = []
  imageids = []
  categories = []

  for i in range(len(images)):
    img_name = images[i]["file_name"]
    images_name.append(img_name)

  for i in range(len(annotation)):
    bboxs.append(annotation[i]["bbox"])
    categories.append(annotation[i]["category_id"])
    imageids.append(annotation[i]["image_id"])

  # 1. Data preprocess (formating)
  coco2map(images_name, bboxs, categories, imageids, groundtruth_dir)


if __name__ == '__main__':
  eval()
