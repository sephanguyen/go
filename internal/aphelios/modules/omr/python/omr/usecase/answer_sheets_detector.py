import json
import os

import logging
import aiohttp
import traceback

ENV = os.getenv('ENV')
ORG = os.getenv('ORG')
METHOD = "http"

async def get_answer_sheet(b64_img):
  timeout = aiohttp.ClientTimeout(total=60)  # seconds

  async with aiohttp.ClientSession(timeout=timeout) as session:
    url = f"{METHOD}://answer-sheet-predictor-default.{ENV}-{ORG}-machine-learning.svc.cluster.local/v1/models/answer_sheet:predict"
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
