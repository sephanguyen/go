import datetime
import json
import os

import click
import cv2
import imgaug as ia
import numpy as np
from imgaug import augmenters as iaa
from imgaug.augmentables.bbs import BoundingBox, BoundingBoxesOnImage
from imgaug.augmentables.polys import Polygon

from coco_format import Annotation, Category, Image, Info, CocoFormatter, cocoformatter_from_dict, \
  cocoformatter_to_dict


class PolygonImage():
  def __init__(self, img: list, polygons: list):
    self.img = img
    self.polygons = polygons


class BboxImage():
  def __init__(self, img: list, bboxs: list, labels: list):
    self.img = img
    self.bboxs = bboxs
    self.labels = labels


def load_polygon(data_dir):
  # 1. Data load
  path2json = f"{data_dir}/result.json"

  try:
    with open(path2json) as json_file:
      data = json.load(json_file)
      img_mp = {}

      coco = cocoformatter_from_dict(data)

      data_images = coco.images
      plgs = []
      for i in range(len(data_images)):
        img_mp[data_images[i].id] = {}
        img_mp[data_images[i].id]["file_name"] = data_images[i].file_name

      data_annotations = coco.annotations
      for i in range(len(data_annotations)):
        seg_float = [float(x) for x in data_annotations[i].segmentation[0]]
        segmentations = np.reshape(seg_float, (-1, 2))

        if "polygons" not in img_mp[data_annotations[i].image_id]:
          img_mp[data_annotations[i].image_id]["polygons"] = []
        img_mp[data_annotations[i].image_id]["polygons"].append(segmentations)

      for _, v in img_mp.items():
        try:
          img = cv2.imread(os.path.join(data_dir, v["file_name"]))
        except IOError:
          print(f'not found {v["file_name"]}')
        plgs.append(PolygonImage(img=img, polygons=v["polygons"]))
  except IOError:
    print("not found result.json file")

  return plgs


def load_bbox(input_dir=""):
  img_mp = {}
  img_label = {}
  bbox_image = []

  result_path = f"{input_dir}/result.json"
  with open(result_path) as label_file:
    label_dict = json.load(label_file)
    coco_format = CocoFormatter.from_dict(label_dict)
    images = coco_format.images
    annotation = coco_format.annotations
    for img in images:
      img_mp[img.file_name] = []
      img_label[img.file_name] = []
    for ann in annotation:
      img_id = ann.image_id
      img_mp[images[img_id].file_name].append(ann.bbox)
      img_label[images[img_id].file_name].append(ann.category_id)

    for img in img_mp.keys():
      bbox_image.append(BboxImage(img=img, bboxs=img_mp[img], labels=img_label))

    return bbox_image


# this will generate augmentation images for question field data (already annotated and generated in coco format)
# coco format contains the result.json file and images folder
# this augmentation will augment data with polygons on images
def generate_image(label_name: str, annotation_dir: str, out_dir: str, is_bubble=False):
  if (is_bubble == False):
    coco = generate_polygon_image(label_name, annotation_dir, out_dir)
  elif (is_bubble == True):
    coco = generate_bbox_image(annotation_dir, out_dir)

  print("finish augmentation...")

  try:
    with open(f"{out_dir}/result.json", 'w') as f:
      data = cocoformatter_to_dict(coco)
      json.dump(data, f, indent=2)
  except IOError:
    print(f'cannot export data to {out_dir}')

def generate_polygon_image(label_name: str, annotation_dir: str, out_dir: str):
  plgs = load_polygon(annotation_dir)

  ia.seed(2)
  aug = iaa.Sequential([
    iaa.PerspectiveTransform((0.01, 0.02)),
    iaa.AddToHueAndSaturation((-20, 20)),
    iaa.LinearContrast((0.95, 1.05), per_channel=0.5),
    iaa.Sometimes(0.75, iaa.Snowflakes())
  ])

  print("starting augmentation...")

  categories = [Category(id=0, name=label_name)]
  info = Info(year=2022, version="1.0.0", description="data with aug", contributor="manabie", url="",
              date_created=datetime.datetime.now())
  coco = CocoFormatter(categories=categories, info=info, images=[], annotations=[])

  for plg in plgs:
    image = plg.img

    iaa_poly = [Polygon(p) for p in plg.polygons]

    psoi = ia.PolygonsOnImage(iaa_poly, shape=image.shape)  # load polygon and image shape
    for _ in range(20):  # gen 20 version in each image
      image_aug, psoi_aug = aug(image=image, polygons=psoi)
      os.makedirs(f'{out_dir}/aug_images', exist_ok=True)
      image_id = coco.get_next_image_id()
      image_res = Image(width=image_aug.shape[0], height=image_aug.shape[1],
                        file_name=f'{out_dir}/aug_images/image_{image_id}.jpg',
                        id=image_id)
      coco.images.append(image_res)
      cv2.imwrite(image_res.file_name, image_aug)

      for p in psoi_aug:
        annotation_id = coco.get_next_annotation_id()
        seg = p.coords.flatten()
        area = p.area
        bbox = p.to_bounding_box().coords.flatten()
        annotation = Annotation(id=annotation_id, image_id=image_id, category_id=coco.categories[0].id,
                                segmentation=[seg], bbox=bbox, area=area, ignore=0, iscrowd=0)
        coco.annotations.append(annotation)
  return coco

def generate_bbox_image(annotation_dir: str, out_dir: str):
  bboxs = load_bbox(annotation_dir)
  ia.seed(1)

  seq = iaa.Sequential([
    iaa.PerspectiveTransform((0.01, 0.02)),
    iaa.AddToHueAndSaturation((-20, 20)),
    iaa.LinearContrast((0.95, 1.05), per_channel=0.5),
    iaa.Sometimes(0.75, iaa.Snowflakes())
  ])
  categories = [Category(id=0, name="background"), Category(id=1, name="confirmed"),
                Category(id=2, name="empty_bubble")]

  info = Info(year=2022, version="1.0.0", description="data with aug", contributor="manabie", url="",
              date_created=datetime.datetime.now())
  coco = CocoFormatter(categories=categories, info=info, images=[], annotations=[])
  for bbox_index in range(len(bboxs)):
    bbox = bboxs[bbox_index]
    img = bbox.img
    bb = bbox.bboxs
    label = bbox.labels[img]

    image = cv2.imread(f"{annotation_dir}/{img}")

    bounding_box = []
    for bubble_index in range(len(bb)):
      x1 = bb[bubble_index][0]
      y1 = bb[bubble_index][1]
      x2 = bb[bubble_index][0] + bb[bubble_index][2]
      y2 = bb[bubble_index][1] + bb[bubble_index][3]
      bounding_box.append(BoundingBox(x1=x1, x2=x2, y1=y1, y2=y2))

    bbs = BoundingBoxesOnImage(bounding_box, shape=image.shape)
    for _ in range(20):

      os.makedirs(f"{out_dir}/images", exist_ok=True)
      image_aug, bbs_aug = seq(image=image, bounding_boxes=bbs)
      image_id = coco.get_next_image_id()
      image_res = Image(width=image_aug.shape[0], height=image_aug.shape[1],
                        file_name=f'{out_dir}/images/image_{image_id}.jpg',
                        id=image_id)
      coco.images.append(image_res)
      cv2.imwrite(image_res.file_name, image_aug)

      for bb_index in range(len(bbs_aug)):
        annotation_id = coco.get_next_annotation_id()
        bbox = bbs_aug[bb_index].coords.flatten()
        coco_bbox = [bbox[0], bbox[1], bbox[2] - bbox[0], bbox[3] - bbox[1]]
        seg = []
        area = 0
        annotation = Annotation(id=annotation_id, image_id=image_id, category_id=label[bb_index]  # get label id.
                                , segmentation=seg, area=area, bbox=coco_bbox, ignore=0, iscrowd=0)
        coco.annotations.append(annotation)
  return coco

@click.command()
@click.option('--label_tag', prompt='label tag name', default="question_field",
              help='Label tag name in strategies field')
@click.option('--is_bubble', is_flag=True,
              prompt='Is augmented for bubble detection?')
@click.option('--input_dir', prompt='input directory', help='COCO folder, which be augmented.')
@click.option('--output_dir', prompt='output directory', help='The directory which store augmentated images')
def run(label_tag, input_dir, output_dir, is_bubble):
  generate_image(label_name=label_tag, annotation_dir=input_dir, out_dir=output_dir, is_bubble=is_bubble)


if __name__ == '__main__':
  run()
