import json
import os

import logging
import aiohttp
import asyncio
import traceback
import numpy as np

from pkg.manabuf_py.aphelios.vision.v1 import vision_pb2_grpc, vision_pb2

SERVICE_NAME = "QUESTION_FIELD_DETECTOR"
ENV = os.getenv('ENV')
ORG = os.getenv('ORG')
METHOD = "http"

async def get_question_field(b64_img):
  timeout = aiohttp.ClientTimeout(total=60)  # seconds

  async with aiohttp.ClientSession(timeout=timeout) as session:
    url = f"{METHOD}://question-field-predictor-default.{ENV}-{ORG}-machine-learning.svc.cluster.local/v1/models/question-field:predict"
    payload = { \
      "instances": [ \
        { \
          "image": { \
            "b64": f"{b64_img}" \
            } \
          }] \
      }

    async with session.post(url, json=payload) as response:
      logging.info("Status:", response.status)
      logging.info("Content-type:", response.headers['content-type'])

      res = await response.text()
      try:
        j = json.loads(res)
        bboxs = j["predictions"]

        return bboxs
      except:
        logging.error(f"traceback:{traceback.print_exception()}")
        return "internal error"

class QuestionFieldDetectorServiceServicer(vision_pb2_grpc.QuestionFieldDetectorServiceServicer):
  def QuestionFieldDetector(self, request, context):
    bboxs_re = []
    scores_re = []
    # Get images
    try:
      b64 = request.b64_img

      # processing
      bboxs = asyncio.run(get_question_field(b64))



      if "internal error" not in bboxs:
        for bbox in bboxs:
          xc, yc, w, h, ag = bbox[:5]
          wx, wy = w / 2 * np.cos(ag), w / 2 * np.sin(ag)
          hx, hy = -h / 2 * np.sin(ag), h / 2 * np.cos(ag)
          p1 = [xc - wx - hx, yc - wy - hy]
          p2 = [xc + wx - hx, yc + wy - hy]
          p3 = [xc + wx + hx, yc + wy + hy]
          p4 = [xc - wx + hx, yc - wy + hy]

          pts = np.array([p1, p2, p3, p4], np.int32)
          score = bbox[5]

          bboxs_re.append(pts)
          scores_re.append(score)
      else:
        bboxs_re.append("internal error")

      return vision_pb2.QuestionFieldResponse(id=request.id, bbox_list=str(bboxs_re),
                                           status=vision_pb2.OmrServiceStatus.Value("SUCCESS"))

    except Exception as e:
      logging.error(f"{SERVICE_NAME} ERROR. Reason: {e} - trace: {traceback.format_exc()}")
      return vision_pb2.QuestionFieldResponse(id=request.id, bbox_list=str(bboxs_re),
                                           status=vision_pb2.OmrServiceStatus.Value('ERROR_IN_QUESTION_FIELD_DETECTOR'))