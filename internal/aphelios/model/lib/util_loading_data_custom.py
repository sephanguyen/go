# USAGE
# python train.py
# import the necessary packages
import torch
import cv2
import numpy as np
import torchvision.datasets as dset

from lib import config
from torchvision import transforms
from torch.utils.data import DataLoader
from sklearn.preprocessing import LabelEncoder

# initialize the list of data (images), class labels, target bounding
# box coordinates, and image paths
print("[INFO] loading dataset...")
data = []
labels = []
bboxes = []
imagePaths = []

path2data = "../data/train"
path2json = "../data/train/result.json"

transform = transforms.Compose([
  transforms.ToTensor()
])


def loading_data():
  # 1. Data load
  path2data = "../data/train"
  path2json = "../data/train/result.json"
  coco_train = dset.CocoDetection(root=path2data,
                                  annFile=path2json, transform=transform)

  train_data_loader = DataLoader(coco_train, batch_size=2, shuffle=True)

  print('Number of samples: ', len(coco_train))

  return train_data_loader, coco_train


# define normalization transforms
transforms = transforms.Compose([
  transforms.ToPILImage(),
  transforms.ToTensor(),
  transforms.Normalize(mean=config.MEAN, std=config.STD)
])
# perform label encoding on the labels
le = LabelEncoder()


def parse_data_object_detection(coco_train, limit=100):
  data = []
  labels = []
  bboxes = []
  image_id = []

  for _index in range(coco_train.__len__()):

    image_tensor, labels_tensor = coco_train.__getitem__(_index)
    image = image_tensor[_index].numpy()  # get the first chanel from image array.
    (h, w) = image.shape
    image = cv2.cvtColor(image, cv2.COLOR_GRAY2RGB)
    data.append(image)

    labels_per_img = []
    bboxes_per_img = []
    image_id_per_img = []

    count = 0
    for label in labels_tensor:
      count = count + 1
      if count <= limit:
        image_id = label["image_id"]
        category_id = label["category_id"]

        bbox = label["bbox"]
        start_x = bbox[0]
        start_y = bbox[1]
        end_x = start_x + bbox[2]
        end_y = start_y + bbox[3]

        image_id_per_img.append(image_id)
        labels_per_img.append(category_id)
        bboxes_per_img.append((start_x, start_y, end_x, end_y))

        labels.append(labels_per_img)
        image_id.append(image_id_per_img)
        bboxes.append(bboxes_per_img)

  data = np.array(data, dtype="float32")
  labels = np.array(labels)
  bboxes = np.array(bboxes, dtype="float32")

  return torch.tensor(data), torch.tensor(image_id), torch.tensor(labels), torch.tensor(bboxes)w

