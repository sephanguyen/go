import logging

import numpy as np
import cv2
import imutils
import io
import traceback

from PIL import Image
from minio import Minio
from imutils.perspective import four_point_transform

from pkg.manabuf_py.aphelios.vision.v1 import vision_pb2_grpc, vision_pb2
from internal.aphelios.modules.omr.python.omr.utils.configure_loader.loader import minio
from internal.aphelios.modules.omr.python.omr.utils.image_processer.processor import convert_to_jpg

SERVICE_NAME = "GET_TOP_VIEW"

class GetTopViewServiceServicer(vision_pb2_grpc.GetTopViewServiceServicer):

  def GetTopView(self, request, context):
    # load images
    buckets, endpoint, access_key, secret_key = minio.load_config()
    client = Minio(endpoint=endpoint, access_key=access_key, secret_key=secret_key, secure=False)

    # Get images
    try:

      res = client.get_object(bucket_name=buckets, object_name=request.image_url)

      data = res.read()
      pil_image = Image.open(io.BytesIO(data))
      image = np.array(pil_image)

      blurred = cv2.GaussianBlur(image, (5, 5), 0)
      edged = cv2.Canny(blurred, 75, 200)

      contours = cv2.findContours(edged.copy(), cv2.RETR_LIST, cv2.CHAIN_APPROX_SIMPLE)

      cnts = imutils.grab_contours(contours)

      # ensure that at least one contour was found
      if len(cnts) > 0:
        # sort the contours according to their size in
        # descending order
        cnts = sorted(cnts, key=cv2.contourArea, reverse=True)
        # loop over the sorted contours
        for c in cnts:
          # approximate the contour
          peri = cv2.arcLength(c, True)
          approx = cv2.approxPolyDP(c, 0.02 * peri, True)
          # if our approximated contour has four points,
          # then we can assume we have found the paper
          if len(approx) == 4:
            docCnt = approx
            break

      docCnt = approx
      warped = four_point_transform(image, docCnt.reshape(4, 2))

      # write to bucket
      sc, img_np = cv2.imencode(".png", warped)
      img = img_np.tobytes()
      image_type = request.image_url.split(".")[1]
      output_url = request.image_url.replace(".", "").replace(image_type, "_topview.png")
      result = client.put_object(bucket_name=buckets, object_name=output_url, data=io.BytesIO(img), length=-1,
                                 part_size=10 * 1024 * 1024)

      logging.info(f"{SERVICE_NAME} result: {result.object_name} - {result.bucket_name}")
      return vision_pb2.GetTopViewResponse(id=request.id, image_url=str(result.object_name),
                                           status=vision_pb2.OmrServiceStatus.Value("SUCCESS"))
    except Exception as e:
      logging.error(f"{SERVICE_NAME} ERROR. Reason: {e} - trace: {traceback.format_exc()}")
      return vision_pb2.GetTopViewResponse(id=request.id, image_url="", status=vision_pb2
                                           .OmrServiceStatus.Value('ERROR_IN_GET_TOP_VIEW'))
