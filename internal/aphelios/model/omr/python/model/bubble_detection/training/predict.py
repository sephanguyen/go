# USAGE
# python predict.py --input dataset/images/face/image_0131.jpg
# import the necessary packages
# from pyimagesearch import config

import sys
sys.path.append("../../../../..")

import pickle
import torch
import cv2
import torchvision
import mimetypes
import numpy

from lib import config
from torchvision.utils import draw_bounding_boxes
from torchvision.models.detection.faster_rcnn import FastRCNNPredictor
from maskRCNN import get_model_object_train, get_transform
from torchvision import transforms

# import imutils


# define normalization transforms
transforms = transforms.Compose([
  transforms.ToPILImage(),
  transforms.ToTensor(),
  transforms.Normalize(mean=config.MEAN, std=config.STD)
])

input = "example.png"


def create_model(num_classes):
  # load Faster RCNN pre-trained model
  model = torchvision.models.detection.fasterrcnn_resnet50_fpn(pretrained=True)

  # get the number of input features
  in_features = model.roi_heads.box_predictor.cls_score.in_features
  # define a new head for the detector with required number of classes
  model.roi_heads.box_predictor = FastRCNNPredictor(in_features, num_classes)
  return model


def run():
  # determine the input file type, but assume that we're working with
  # single input image
  filetype = mimetypes.guess_type(input)[0]
  imagePaths = [input]
  # if the file type is a text file, then we need to process *multiple*
  # images
  if "text/plain" == filetype:
    # load the image paths in our testing file
    imagePaths = open(imagePaths).read().strip().split("\n")

  print("[INFO] loading object detector...")
  model = torch.load(config.MODEL_PATH).to(config.DEVICE)
  model.eval()
  le = pickle.loads(open(config.LE_PATH, "rb").read())

  # regression model
  for imagePath in imagePaths:
    # load the image, copy it, swap its colors channels, resize it, and
    # bring its channel dimension forward
    image = cv2.imread(imagePath)
    orig = image.copy()
    image = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
    image = cv2.cvtColor(image, cv2.COLOR_GRAY2RGB)
    # image = cv2.resize(image, (224, 224))
    # image = image.transpose((2, 0, 1))
    # convert image to PyTorch tensor, normalize it, flash it to the
    # current device, and add a batch dimension
    # image = torch.from_numpy(image)
    image = transforms(image).to(config.DEVICE)
    image = image.unsqueeze(0)

    # predict the bounding box of the object along with the class
    # label
    (boxPreds, labelPreds) = model(image)
    (startX, startY, endX, endY) = boxPreds[0]
    # determine the class label with the largest predicted
    # probability
    labelPreds = torch.nn.Softmax(dim=-1)(labelPreds)
    i = labelPreds.argmax(dim=-1).cpu()
    # label = le.inverse_transform(i)[0]

    # resize the original image such that it fits on our screen, and
    # grab its dimensions
    # orig = imutils.resize(orig, width=600)
    (h, w) = orig.shape[:2]
    # scale the predicted bounding box coordinates based on the image
    # dimensions
    startX = int(startX * w)
    startY = int(startY * h)
    endX = int(endX * w)
    endY = int(endY * h)
    # draw the predicted bounding box and class label on the image
    # y = startY - 10 if startY - 10 > 10 else startY + 10
    # cv2.putText(orig, str(i), (startX, y), cv2.FONT_HERSHEY_SIMPLEX,
    # 						0.65, (0, 255, 0), 2)
    cv2.rectangle(orig, (startX, startY), (endX, endY), 0, 2)
    # show the output image
    cv2.imshow("Output", orig)
    cv2.waitKey(0)


def rcnn():
  transforms = get_transform(False)

  image = cv2.imread(input)
  image = cv2.cvtColor(image, cv2.COLOR_BGR2GRAY)
  image = cv2.cvtColor(image, cv2.COLOR_GRAY2RGB)

  model = get_model_object(2, 0)
  checkpoint = torch.load(f'{config.BASE_OUTPUT}/best_model.pth', map_location=config.DEVICE)
  model.load_state_dict(checkpoint['model_state_dict'])
  model.to(config.DEVICE).eval()


  print("LOADING DONE")

  image_tensor = torch.tensor(image).permute(2, 0, 1)

  image_ = torch.from_numpy(numpy.array([image_tensor.numpy()])).to(config.DEVICE)
  image_nor = transforms(image_).to(config.DEVICE)

  output = model(image_nor)

  print("PREDICT DONE")

  score_threshold = 0.5
  omr_box = draw_bounding_boxes(image_[0], boxes=output[0]['boxes'][output[0]['scores'] > score_threshold], width=4)\
    .to("cpu").permute(1, 2, 0).numpy()

  cv2.imwrite("result.jpg", omr_box)
  cv2.waitKey(0)

if __name__ == '__main__':
  rcnn()
