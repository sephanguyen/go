# import the necessary packages
import torch
import os

CLASSES = 2

# flag
DEVICE = "cuda" if torch.cuda.is_available() else "cpu"
PIN_MEMORY = True if DEVICE == "cuda" else False

# Answer sheets size:
WIDTH = 960
HEIGHT = 1280

# specify ImageNet mean and standard deviation
MEAN = [0.485, 0.456, 0.406]
STD = [0.229, 0.224, 0.225]
BATCH_SIZE = 1

# initialize our initial learning rate, number of epochs to train
# for, and the batch size
INIT_LR = 0.001
NUM_EPOCHS = 50

# specify the loss weights
LABELS = 1.0
BBOX = 1.0

# define the base path to the input dataset and then use it to derive
# the path to the input images and annotation CSV files
BASE_PATH = "dataset"
IMAGES_PATH = os.path.sep.join([BASE_PATH, "images"])
ANNOTS_PATH = os.path.sep.join([BASE_PATH, "annotations"])
# define the path to the base output directory
