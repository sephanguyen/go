import json
import os
import logging
import aiohttp
import asyncio
import traceback

from pkg.manabuf_py.aphelios.vision.v1 import vision_pb2_grpc, vision_pb2

SERVICE_NAME = "BUBBLE_DETECTOR"
ENV = os.getenv('ENV')
ORG = os.getenv('ORG')
METHOD = "http"


async def get_bubble_answer(b64_img):
  timeout = aiohttp.ClientTimeout(total=3600)  # seconds

  async with aiohttp.ClientSession(timeout=timeout) as session:
    url = f"{METHOD}://bubble-predictor-default.{ENV}-{ORG}-machine-learning.svc.cluster.local/v1/models/bubble-model:predict"
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
        bboxs = j["predictions"]["boxes"]
        labels = j["predictions"]["labels"]
        scores = j["predictions"]["scores"]

        return bboxs, labels, scores
      except:
        logging.error(f"traceback:{traceback.print_exception()}")
        return "internal error", "internal error", "internal error"

class BubbleDetectorServiceServicer(vision_pb2_grpc.BubbleDetectorServiceServicer):
  def BubbleDetector(self, request, context):
    # Get images
    bboxs_re = []
    try:
      b64 = request.b64_img
      # processing
      bboxs, labels, scores = asyncio.run(get_bubble_answer(b64))

      if "intern1al error" not in bboxs:
        for bbox in bboxs:
          x1, y1, x2, y2 = bbox[:4]
          p1 = (int(x1), int(y1))
          p2 = (int(x2), int(y2))
          pts = [p1, p2]

          bboxs_re.append(pts)
      else:
        bboxs_re.append("internal error")

      return vision_pb2.BubbleDetectionResponse(id=request.id, bbox_list=str(bboxs_re),
                                                status=vision_pb2.OmrServiceStatus.Value("SUCCESS"))

    except Exception as e:
      logging.error(f"{SERVICE_NAME} ERROR. Reason: {e} - trace: {traceback.format_exc()}")
      return vision_pb2.BubbleDetectionResponse(id=request.id, bbox_list=str(bboxs_re),
                                                status=vision_pb2.OmrServiceStatus.Value(
                                                  'ERROR_IN_BUBBLE_DETECTOR'))
