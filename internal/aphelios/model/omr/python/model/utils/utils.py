import json


def loading_data(path2json="/home/vongho/Documents/vongho/src/omr_pytorch/data_for_trainning/4corner/result.json"):
  """
  Loading coco data format, then return list images, segmentation, label and categories
  Args:
    path2json: json file, result of annotation in coco format.

  Returns:
    images (list)
    segments (list)
    labels (list)
    categories (list)
  """
  images = []
  segments = []
  labels = []
  categories = []

  # 1. Data load
  with open(path2json) as json_file:
    data = json.load(json_file)

    # Label name
    for i in range(len(data["categories"])):
      categories.append(data["categories"][i]["name"])

    # Get images list:
    for i in range(len(data["images"])):
      images.append(data["images"][i]["file_name"])

      # Get segments list:
      segments.append(data["annotations"][i]["segmentation"][0])

      # Get labels:
      labels.append(data["annotations"][i]["category_id"])

  print("DONE")
  return images, segments, labels, categories
