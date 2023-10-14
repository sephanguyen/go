import datetime
from typing import List, Any, TypeVar, Callable, Type, cast

import dateutil.parser

T = TypeVar("T")


def from_int(x: Any) -> int:
  return x


def from_list(f: Callable[[Any], T], x: Any) -> List[T]:
  return [f(y) for y in x]


def from_float(x: Any) -> float:
  return float(x)


def to_float(x: Any) -> float:
  return float(x)


def from_str(x: Any) -> str:
  return x


def from_datetime(x: Any) -> datetime:
  return dateutil.parser.parse(x)


def to_class(c: Type[T], x: Any) -> dict:
  assert isinstance(x, c)
  return cast(Any, x).to_dict()


class Annotation:
  id: int
  image_id: int
  category_id: int
  segmentation: List[List[float]]
  bbox: List[float]
  ignore: int
  iscrowd: int
  area: float

  def __init__(self, id: int, image_id: int, category_id: int, segmentation: List[List[float]], bbox: List[float],
               ignore: int, iscrowd: int, area: float) -> None:
    self.id = id
    self.image_id = image_id
    self.category_id = category_id
    self.segmentation = segmentation
    self.bbox = bbox
    self.ignore = ignore
    self.iscrowd = iscrowd
    self.area = area

  @staticmethod
  def from_dict(obj: Any) -> 'Annotation':
    assert isinstance(obj, dict)
    annotation_id = from_int(obj.get("id"))
    image_id = from_int(obj.get("image_id"))
    category_id = from_int(obj.get("category_id"))
    segmentation = from_list(lambda x: from_list(from_float, x), obj.get("segmentation"))
    bbox = from_list(from_float, obj.get("bbox"))
    ignore = from_int(obj.get("ignore"))
    iscrowd = from_int(obj.get("iscrowd"))
    area = from_float(obj.get("area"))
    return Annotation(annotation_id, image_id, category_id, segmentation, bbox, ignore, iscrowd, area)

  def to_dict(self) -> dict:
    result: dict = {}
    result["id"] = from_int(self.id)
    result["image_id"] = from_int(self.image_id)
    result["category_id"] = from_int(self.category_id)
    result["segmentation"] = from_list(lambda x: from_list(to_float, x), self.segmentation)
    result["bbox"] = from_list(to_float, self.bbox)
    result["ignore"] = from_int(self.ignore)
    result["iscrowd"] = from_int(self.iscrowd)
    result["area"] = to_float(self.area)
    return result


class Category:
  id: int
  name: str

  def __init__(self, id: int, name: str) -> None:
    self.id = id
    self.name = name

  @staticmethod
  def from_dict(obj: Any) -> 'Category':
    assert isinstance(obj, dict)
    category_id = from_int(obj.get("id"))
    name = from_str(obj.get("name"))
    return Category(category_id, name)

  def to_dict(self) -> dict:
    result: dict = {}
    result["id"] = from_int(self.id)
    result["name"] = from_str(self.name)
    return result


class Image:
  width: int
  height: int
  id: int
  file_name: str

  def __init__(self, width: int, height: int, id: int, file_name: str) -> None:
    self.width = width
    self.height = height
    self.id = id
    self.file_name = file_name

  @staticmethod
  def from_dict(obj: Any) -> 'Image':
    assert isinstance(obj, dict)
    width = from_int(obj.get("width"))
    height = from_int(obj.get("height"))
    image_id = from_int(obj.get("id"))
    file_name = from_str(obj.get("file_name"))
    return Image(width, height, image_id, file_name)

  def to_dict(self) -> dict:
    result: dict = {}
    result["width"] = from_int(self.width)
    result["height"] = from_int(self.height)
    result["id"] = from_int(self.id)
    result["file_name"] = from_str(self.file_name)
    return result


class Info:
  year: int
  version: str
  description: str
  contributor: str
  url: str
  date_created: datetime

  def __init__(self, year: int, version: str, description: str, contributor: str, url: str,
               date_created: datetime) -> None:
    self.year = year
    self.version = version
    self.description = description
    self.contributor = contributor
    self.url = url
    self.date_created = date_created

  @staticmethod
  def from_dict(obj: Any) -> 'Info':
    assert isinstance(obj, dict)
    year = from_int(obj.get("year"))
    version = from_str(obj.get("version"))
    description = from_str(obj.get("description"))
    contributor = from_str(obj.get("contributor"))
    url = from_str(obj.get("url"))
    date_created = from_datetime(obj.get("date_created"))
    return Info(year, version, description, contributor, url, date_created)

  def to_dict(self) -> dict:
    result: dict = {}
    result["year"] = from_int(self.year)
    result["version"] = from_str(self.version)
    result["description"] = from_str(self.description)
    result["contributor"] = from_str(self.contributor)
    result["url"] = from_str(self.url)
    result["date_created"] = self.date_created.isoformat()
    return result


class CocoFormatter:
  images: List[Image]
  categories: List[Category]
  annotations: List[Annotation]
  info: Info

  def __init__(self, images: List[Image], categories: List[Category], annotations: List[Annotation],
               info: Info) -> None:
    self.images = images
    self.categories = categories
    self.annotations = annotations
    self.info = info

  def get_next_image_id(self):
    return len(self.images)

  def get_next_annotation_id(self):
    return len(self.annotations)

  @staticmethod
  def from_dict(obj: Any) -> 'CocoFormatter':
    assert isinstance(obj, dict)
    images = from_list(Image.from_dict, obj.get("images"))
    categories = from_list(Category.from_dict, obj.get("categories"))
    annotations = from_list(Annotation.from_dict, obj.get("annotations"))
    info = Info.from_dict(obj.get("info"))
    return CocoFormatter(images, categories, annotations, info)

  def to_dict(self) -> dict:
    result: dict = {}
    result["images"] = from_list(lambda x: to_class(Image, x), self.images)
    result["categories"] = from_list(lambda x: to_class(Category, x), self.categories)
    result["annotations"] = from_list(lambda x: to_class(Annotation, x), self.annotations)
    result["info"] = to_class(Info, self.info)
    return result


def cocoformatter_from_dict(s: Any) -> CocoFormatter:
  return CocoFormatter.from_dict(s)


def cocoformatter_to_dict(x: CocoFormatter) -> Any:
  return to_class(CocoFormatter, x)
