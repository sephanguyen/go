import logging

import numpy as np
import cv2
import imutils
import io
import traceback

from PIL import Image
from minio import Minio

from pkg.manabuf_py.aphelios.vision.v1 import vision_pb2_grpc, vision_pb2
from internal.aphelios.modules.omr.python.omr.utils.configure_loader.loader import minio

SERVICE_NAME = "ID_QUESTION_DETECTOR"

class IdQuestionDetectorServiceServicer(vision_pb2_grpc.IdQuestionDetectorServiceServicer):

  def IdQuestionDetector(self, request, context):
    # load images
    buckets, endpoint, access_key, secret_key = minio.load_config()
    client = Minio(endpoint=endpoint, access_key=access_key, secret_key=secret_key, secure=False)

    # Get images
    try:

      res = client.get_object(bucket_name=buckets, object_name=request.image_url)

      data = res.read()
      pil_image = Image.open(io.BytesIO(data))
      warped = np.array(pil_image)
      warped = cv2.cvtColor(warped, cv2.COLOR_BGR2GRAY)

      # processing

      # Finding question contours
      bin_image = cv2.adaptiveThreshold(warped, 255, cv2.ADAPTIVE_THRESH_GAUSSIAN_C, cv2.THRESH_BINARY_INV, 77, 14)
      cnts = cv2.findContours(bin_image.copy(), cv2.RETR_LIST, cv2.CHAIN_APPROX_SIMPLE)
      docCnts = imutils.grab_contours(cnts)

      black = np.zeros(bin_image.shape, dtype="uint8")

      ############################################################
      warped2 = warped.copy()
      questionCnts = []
      # loop over the contours
      for c in docCnts:
        # compute the bounding box of the contour, then use the
        # bounding box to derive the aspect ratio
        (x, y, w, h) = cv2.boundingRect(c)
        im = cv2.rectangle(warped2, (x, y), (x + w, y + h), (0, 100, 100, 100), 1)
        ar = w / float(h)
        # in order to label the contour as a question, region
        # should be sufficiently wide, sufficiently tall, and
        # have an aspect ratio approximately equal to 1
        if y < 800 and w >= 20 and h >= 20 and ar >= 0.5 and ar <= 1.5:  # config manually
          questionCnts.append(c)
      # filter question bubble
      answer_zone = cv2.drawContours(black.copy(), questionCnts, -1, (255, 255, 0), 1)

      # write to bucket
      # 1. answer_zone
      sc, img_np = cv2.imencode(".png", answer_zone)
      img = img_np.tobytes()
      image_type = request.image_url.split(".")[1]
      output_url_ans = request.image_url.replace(".", "").replace(image_type, "_anszone.png")

      result_ans = client.put_object(bucket_name=buckets, object_name=output_url_ans, data=io.BytesIO(img), length=-1,
                                     part_size=10 * 1024 * 1024)
      logging.info(f"{SERVICE_NAME} result answer zone: {result_ans.object_name} - {result_ans.bucket_name}")

      # 2. bin_image
      sc, img_bin_np = cv2.imencode(".png", bin_image)
      img_bin = img_bin_np.tobytes()
      image_type = request.image_url.split(".")[1]
      output_url_bin = request.image_url.replace(".", "").replace(image_type, "_bin.png")

      result_bin = client.put_object(bucket_name=buckets, object_name=output_url_bin, data=io.BytesIO(img_bin),
                                     length=-1,
                                     part_size=10 * 1024 * 1024)
      logging.info(f"{SERVICE_NAME} result bin: {result_bin.object_name} - {result_bin.bucket_name}")

      return vision_pb2.IdQuestionResponse(id=request.id, ans_img_url=str(result_ans.object_name),
                                           bin_img_url=str(result_bin.object_name),
                                           status=vision_pb2.OmrServiceStatus.Value("SUCCESS"))
    except Exception as e:
      logging.error(f"{SERVICE_NAME} ERROR. Reason: {e} - trace: {traceback.format_exc()}")
      return vision_pb2.IdQuestionResponse(id=request.id, image_url="", status=vision_pb2
                                           .OmrServiceStatus.Value('ERROR_IN_ID_QUESTION_DETECTOR'))
