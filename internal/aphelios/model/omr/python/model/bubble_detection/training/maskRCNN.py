# USAGE
# python train.py
# import the necessary packages
import sys
import os
import json
import logging
import time

import numpy
import torch
import mlflow
import torchvision
import torchvision.datasets as dset


sys.path.append("../../../../..")

from lib import config
from tqdm import tqdm
from torchvision import transforms

from lib.dataloading import MaskRNNDataset
from lib.utils import Averager, SaveBestModel, save_model, save_loss_plot
from lib.util_loading_data_custom import parse_data_object_detection
from lib.util_loading_data_custom import loading_data, parse_data_object_detection

from torchvision import transforms as T
from torchvision.models.detection.faster_rcnn import FastRCNNPredictor
from torchvision.models.detection import FasterRCNN_ResNet50_FPN_Weights
from torchvision.models import ResNet50_Weights
from torch.utils.data import DataLoader

from sklearn.preprocessing import LabelEncoder


# initialize the list of data (images), class labels, target bounding
# box coordinates, and image paths
logging.debug("[INFO] loading dataset...")
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

  logging.debug('Number of samples: ', len(coco_train))

  return train_data_loader, coco_train


# perform label encoding on the labels
le = LabelEncoder()


def get_model_object(num_classes, trainable):
  # load a model pre-trained on COCO
  model = torchvision.models.detection.fasterrcnn_resnet50_fpn(weights=FasterRCNN_ResNet50_FPN_Weights.COCO_V1,
                                                               weights_backbone=ResNet50_Weights.IMAGENET1K_V1,
                                                               trainable_backbone_layers=trainable
                                                               )

  # get the number of input features
  in_features = model.roi_heads.box_predictor.cls_score.in_features
  # define a new head for the detector with required number of classes
  model.roi_heads.box_predictor = FastRCNNPredictor(in_features, num_classes)

  return model


def get_transform(train):
  transforms = []
  # transforms.append(T.PILToTensor())
  transforms.append(T.ConvertImageDtype(torch.float))
  if train:
    transforms.append(T.RandomHorizontalFlip(0.5))
  return T.Compose(transforms)


transforms = get_transform(True)


# function for running training iterations
def train(train_data_loader, model):
  logging.debug('Training')
  global train_itr
  global train_loss_list

  # initialize tqdm progress bar
  prog_bar = tqdm(train_data_loader, total=len(train_data_loader))

  for i, data in enumerate(prog_bar):
    optimizer.zero_grad()
    images, targets = data

    images = list(image.to(config.DEVICE) for image in images)
    _target = []

    d = {}
    d["boxes"] = targets["boxes"][i].to(config.DEVICE)
    d["labels"] = targets["labels"][i].to(config.DEVICE)
    _target.append(d)


    loss_dict = model(images, _target)

    losses = sum(loss for loss in loss_dict.values())
    loss_value = losses.item()
    train_loss_list.append(loss_value)
    train_loss_hist.send(loss_value)
    losses.backward()
    optimizer.step()
    train_itr += 1

    # update the loss value beside the progress bar for each iteration
    prog_bar.set_description(desc=f"Loss: {loss_value:.4f}")
  return train_loss_list


# function for running validation iterations
def validate(valid_data_loader, model):
  logging.debug('Validating')
  global val_itr
  global val_loss_list

  # initialize tqdm progress bar
  prog_bar = tqdm(valid_data_loader, total=len(valid_data_loader))

  for i, data in enumerate(prog_bar):
    images, targets = data

    images = list(image.to(config.DEVICE) for image in images)
    _target = []
    d = {}
    d["boxes"] = targets["boxes"][i].to(config.DEVICE)
    d["labels"] = targets["labels"][i].to(config.DEVICE)
    _target.append(d)

    with torch.no_grad():
      loss_dict = model(images, _target)
    losses = sum(loss for loss in loss_dict.values())
    loss_value = losses.item()
    val_loss_list.append(loss_value)
    val_loss_hist.send(loss_value)
    val_itr += 1
    # update the loss value beside the progress bar for each iteration
    prog_bar.set_description(desc=f"Loss: {loss_value:.4f}")
  return val_loss_list


if __name__ == '__main__':
  train_data_loader, coco_train = loading_data()
  logging.debug("LOADING DONE")

  logging.debug("PARSING IMAGE ...")
  trainImages, trainImageIds, trainLabels, trainBBoxes = parse_data_object_detection(coco_train, limit=2000)
  logging.debug("PARSING DONE")

  logging.debug("get data set")
  trainDS = MaskRNNDataset((trainImages, trainImageIds, trainLabels, trainBBoxes), transforms=None)
  # calculate steps per epoch for training and validation set
  trainSteps = len(trainDS) // config.BATCH_SIZE

  # create data loaders
  trainLoader = DataLoader(trainDS, batch_size=config.BATCH_SIZE,
                           shuffle=True, num_workers=os.cpu_count(), pin_memory=config.PIN_MEMORY)

  #####################################################
  # get the model using our helper function

  with mlflow.start_run() as run:
    num_classes = 3
    model = get_model_object(num_classes, 2)

    # move model to the right device
    model.to(config.DEVICE)

    # construct an optimizer
    params = [p for p in model.parameters() if p.requires_grad]
    momentum = 0.9
    weight_decay = 0.0005

    step_size = 3
    gamma = 0.1
    optimizer = torch.optim.SGD(params, lr=config.INIT_LR,
                                momentum=momentum, weight_decay=weight_decay)
    # and a learning rate scheduler
    lr_scheduler = torch.optim.lr_scheduler.StepLR(optimizer,
                                                   step_size=step_size,
                                                   gamma=gamma)
    # Log parram
    mlflow.log_params({
      "momentum": momentum,
      "weight_decay": weight_decay,
      "step_size": step_size,
      "gamma": gamma
    })

    train_loss_hist = Averager()
    train_loss_list = []
    train_loss_list = []

    val_loss_hist = Averager()
    val_loss_list = []

    train_itr = 1
    val_itr = 1

    # initialize SaveBestModel class
    save_best_model = SaveBestModel()

    # let's train it for 10 epochs

    # start the training epochs
    for epoch in range(config.NUM_EPOCHS):
      torch.cuda.empty_cache()
      logging.debug(f"\nEPOCH {epoch + 1} of {config.NUM_EPOCHS}")
      # reset the training and validation loss histories for the current epoch
      train_loss_hist.reset()
      val_loss_hist.reset()
      # start timer and carry out training and validation
      start = time.time()
      train_loss = train(trainLoader, model)
      val_loss = validate(trainLoader, model)
      logging.debug(f"Epoch #{epoch + 1} train loss: {train_loss_hist.value:.3f}")
      logging.debug(f"Epoch #{epoch + 1} validation loss: {val_loss_hist.value:.3f}")
      end = time.time()
      logging.debug(f"Took {((end - start) / 60):.3f} minutes for epoch {epoch}")
      # save the best model till now if we have the least loss in the...
      # ... current epoch
      save_best_model(
        val_loss_hist.value, epoch, model, optimizer
      )
      
      logging.debug("[INFO] saving object detector model...")
      #Log metrics
      mlflow.log_metrics({
        "train_loss": train_loss,
        "val_loss": val_loss
      }, epoch)


    logging.debug("That's it!")
    # serialize the model to disk
