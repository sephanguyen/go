import numpy as np
import torch
import torchvision
from torchvision import transforms as T
from torchvision.models import ResNet50_Weights
from torchvision.models.detection import FasterRCNN_ResNet50_FPN_Weights
from torchvision.models.detection.faster_rcnn import FastRCNNPredictor


def convert_to_jpg(pil_image):
  open_cv_image = np.array(pil_image)
  open_cv_image = open_cv_image[:, :, ::-1].copy()
  image = open_cv_image

  return image


def get_model_object_serve(num_classes, trainable):
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
