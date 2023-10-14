import logging

import numpy as np
import cv2
import imutils
import io
import traceback

from PIL import Image
from minio import Minio
from imutils import contours as ct

from pkg.manabuf_py.aphelios.vision.v1 import vision_pb2_grpc, vision_pb2
from internal.aphelios.modules.omr.python.omr.utils.configure_loader.loader import minio
from internal.aphelios.modules.omr.python.omr.utils.image_processer.processor import convert_to_jpg

SERVICE_NAME = "ANSWER-DETECTOR"

class AnswerDetectorServiceServicer(vision_pb2_grpc.AnswerDetectorServiceServicer):
  def AnswerDetector(self, request, context):
    # load images
    buckets, endpoint, access_key, secret_key = minio.load_config()
    client = Minio(endpoint=endpoint, access_key=access_key, secret_key=secret_key, secure=False)

    # Get images
    try:

      res = client.get_object(bucket_name=buckets, object_name=request.top_view_img_url)
      data = res.read()
      pil_image = Image.open(io.BytesIO(data))
      top_view_img = convert_to_jpg(pil_image)
      top_view_img = cv2.cvtColor(top_view_img, cv2.COLOR_BGR2GRAY)

      res = client.get_object(bucket_name=buckets, object_name=request.bin_img_url)
      data = res.read()
      pil_image = Image.open(io.BytesIO(data))
      nimg = np.array(pil_image)
      bin_img = nimg

      res = client.get_object(bucket_name=buckets, object_name=request.ans_img_url)
      data = res.read()
      pil_image = Image.open(io.BytesIO(data))
      nimg = np.array(pil_image)
      anszone_img = nimg

      # processing
      cnts = cv2.findContours(bin_img.copy(), cv2.RETR_LIST, cv2.CHAIN_APPROX_SIMPLE)
      docCnts = imutils.grab_contours(cnts)
      questionCnts = []
      # loop over the contours
      for c in docCnts:
        # compute the bounding box of the contour, then use the
        # bounding box to derive the aspect ratio
        (x, y, w, h) = cv2.boundingRect(c)
        ar = w / float(h)
        # in order to label the contour as a question, region
        # should be sufficiently wide, sufficiently tall, and
        # have an aspect ratio approximately equal to 1
        if y < 800 and w >= 20 and h >= 20 and ar >= 0.5 and ar <= 1.5:  # config manually
          questionCnts.append(c)


      black = np.zeros(bin_img.shape, dtype="uint8")

      # filter external question buble
      cnts = cv2.findContours(anszone_img, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_SIMPLE)
      cnts = imutils.grab_contours(cnts)

      questionCnts = ct.sort_contours(cnts, method="top-to-bottom")[0]
      bubble_choice = {}
      bubble_total = {}
      list_choice = []
      list_unchoice = []

      result_3cnel = cv2.cvtColor(top_view_img.copy(), cv2.COLOR_GRAY2RGB)
      # result_3cnel = top_view_img.copy()

      # each rows has 10 possible answers, to loop over the
      # question in batches of 10
      for (q, i) in enumerate(np.arange(0, len(questionCnts), 10)):
        # sort the contours for the current question from
        # left to right, then initialize the index of the
        # bubbled answer
        sorted_cnts = ct.sort_contours(questionCnts[i:i + 10], method="left-to-right")[0]
        bubbled = None

        # loop over the sorted contours
        for (j, c) in enumerate(sorted_cnts):
          # construct a mask that reveals only the current
          # "bubble" for the question
          mask = np.zeros(bin_img.shape, dtype="uint8")
          bubble_filter = cv2.drawContours(mask.copy(), [c], -1, (255, 255, 0, 100), -1)

          # apply the mask to the thresholded image, then
          # count the number of non-zero pixels in the
          # bubble area
          mask = cv2.bitwise_and(bin_img, bubble_filter.copy(), mask=bin_img)
          total = cv2.countNonZero(mask)

          im3_cp = bubble_filter.copy()

          cnts3 = cv2.findContours(im3_cp, cv2.RETR_EXTERNAL, cv2.CHAIN_APPROX_NONE)
          cnts3_ = imutils.grab_contours(cnts3)

          if total > 500:
            bubble_choice[i + j] = 1
            bubble_total[i + j] = total
            list_choice.append(i + j)
            cv2.drawContours(result_3cnel, cnts3_, -1, (0, 255, 0), 3)
          else:
            bubble_choice[i + j] = 0
            bubble_total[i + j] = total
            list_unchoice.append(i + j)
            cv2.drawContours(result_3cnel, cnts3_, -1, (255, 255, 0), 3)

      # write to bucket
      sc, img_np = cv2.imencode(".png", result_3cnel)
      img = img_np.tobytes()
      image_type = request.top_view_img_url.split(".")[1]
      output_url = request.top_view_img_url.replace(".", "").replace(image_type, "_result.png")
      result = client.put_object(bucket_name=buckets, object_name=output_url, data=io.BytesIO(img), length=-1,
                                 part_size=10 * 1024 * 1024)

      logging.info(f"{SERVICE_NAME} result: {result.object_name} - {result.bucket_name}")
      return vision_pb2.AnswerDetectorResponse(id=request.id, image_url=str(result.object_name),
                                               bubble_choice=list_choice,
                                               status=vision_pb2.OmrServiceStatus.Value("SUCCESS"))
    except Exception as e:
      logging.error(f"{SERVICE_NAME} ERROR. Reason: {e} - trace: {traceback.format_exc()}")
      return vision_pb2.AnswerDetectorResponse(id=request.id, image_url="",
                                               status=vision_pb2.OmrServiceStatus.Value('ERROR_IN_ANSWER_DETECTOR'))
