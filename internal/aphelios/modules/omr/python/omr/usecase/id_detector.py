import json
import os

import logging
import aiohttp
import traceback

SERVICE_NAME = "OMR_GRADING"
ENV = os.getenv('ENV')
ORG = os.getenv('ORG')
METHOD = "http"


async def get_id(b64_img):
  timeout = aiohttp.ClientTimeout(total=3600)  # seconds

  async with aiohttp.ClientSession(timeout=timeout) as session:
    url = f"{METHOD}://ocr-predictor-default.{ENV}-{ORG}-machine-learning.svc.cluster.local/v1/models/ocr:predict"
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
        id = j["predictions"]

        return id
      except:
        logging.error(f"traceback:{traceback.print_exception()}")
        return "internal error"