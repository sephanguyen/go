import cv2
import base64
import numpy as np

from PIL import Image
from io import BytesIO


def convert_to_jpg(pil_image):
  open_cv_image = np.array(pil_image)
  open_cv_image = open_cv_image[:, :, ::-1].copy()
  image = open_cv_image

  return image


def img_to_base64(pil_image):
  im_file = BytesIO()
  pil_image.save(im_file, format="JPEG")
  im_bytes = im_file.getvalue()  # im_bytes: image in binary format.
  im_b64 = base64.b64encode(im_bytes)

  str_b64 = str(im_b64)[2:-1]  # remove header encode string.

  return str_b64


def b64_to_img(b64):
  raw_img_data = base64.b64decode(b64)
  input_image = Image.open(BytesIO(raw_img_data))
  image = convert_to_jpg(input_image)

  return image


def order_points(pts):
  rect = np.zeros((4, 2), dtype="float32")

  s = pts.sum(axis=1)
  rect[0] = pts[np.argmin(s)]
  rect[2] = pts[np.argmax(s)]

  diff = np.diff(pts, axis=1)
  rect[1] = pts[np.argmin(diff)]
  rect[3] = pts[np.argmax(diff)]

  return rect


def four_point_transform(image, pts):
  p1, p2, p3, p4 = pts
  pts = np.array([
    p1,
    p2,
    p3,
    p4], dtype="float32")
  rect = order_points(pts)
  (tl, tr, br, bl) = rect

  widthA = np.sqrt(((br[0] - bl[0]) ** 2) + ((br[1] - bl[1]) ** 2))
  widthB = np.sqrt(((tr[0] - tl[0]) ** 2) + ((tr[1] - tl[1]) ** 2))
  maxWidth = max(int(widthA), int(widthB))

  heightA = np.sqrt(((tr[0] - br[0]) ** 2) + ((tr[1] - br[1]) ** 2))
  heightB = np.sqrt(((tl[0] - bl[0]) ** 2) + ((tl[1] - bl[1]) ** 2))
  maxHeight = max(int(heightA), int(heightB))

  dst = np.array([
    [0, 0],
    [maxWidth - 1, 0],
    [maxWidth - 1, maxHeight - 1],
    [0, maxHeight - 1]], dtype="float32")

  # compute the perspective transform matrix and then apply it
  M = cv2.getPerspectiveTransform(rect, dst)
  warped = cv2.warpPerspective(image, M, (maxWidth, maxHeight))

  # return the warped image
  return warped


def rbox2bbox(rbox):
  xc, yc, w, h, ag = rbox[:5]
  wx, wy = w / 2 * np.cos(ag), w / 2 * np.sin(ag)
  hx, hy = -h / 2 * np.sin(ag), h / 2 * np.cos(ag)
  p1 = [xc - wx - hx, yc - wy - hy]
  p2 = [xc + wx - hx, yc + wy - hy]
  p3 = [xc + wx + hx, yc + wy + hy]
  p4 = [xc - wx + hx, yc - wy + hy]

  pts = np.array([p1, p2, p3, p4], np.int32)
  score = rbox[5]

  return score, pts

def deblur_image(img):
  kernel = np.array([[-1, -1, -1], [-1, 9, -1], [-1, -1, -1]])
  deblur_img = cv2.filter2D(img, -1, kernel)
  gause_blur = cv2.GaussianBlur(deblur_img, (3, 3), 0)
  return gause_blur
