import json
import os

import cv2
import logging
import asyncio
import traceback
import pandas as pd

from PIL import Image
from pkg.manabuf_py.aphelios.vision.v1 import vision_pb2_grpc, vision_pb2
from internal.aphelios.modules.omr.python.omr.usecase.question_field_detector import get_question_field
from internal.aphelios.modules.omr.python.omr.usecase.bubble_detector import get_bubble_answer
from internal.aphelios.modules.omr.python.omr.usecase.answer_sheets_detector import get_answer_sheet
from internal.aphelios.modules.omr.python.omr.usecase.id_detector import get_id
from internal.aphelios.modules.omr.python.omr.utils.image_processer.processor import img_to_base64, rbox2bbox, \
  b64_to_img, four_point_transform, deblur_image

SERVICE_NAME = "OMR_GRADING"
ENV = os.getenv('ENV')
ORG = os.getenv('ORG')
METHOD = "http"


async def grading_omr(b64_img):
  raw_img = b64_to_img(b64_img)
  img = raw_img.copy()
  grading = {}
  max_score = 0

  # 1. ANSWER SHEETS
  logging.info("1. Answer sheets")
  ans_rboxs = await get_answer_sheet(b64_img)
  re_ans_bbox = []

  for rbox in ans_rboxs:
    score, ans_bboxs = rbox2bbox(rbox)
    if score >= max_score:
      re_ans_bbox = ans_bboxs

  ans_img = four_point_transform(img, re_ans_bbox)

  # 2. QUESTION FIELD
  logging.info("2. Question field")
  img_re = ans_img.copy()
  ans_b64 = img_to_base64(Image.fromarray(ans_img))

  img2 = img_re.copy()

  field_rboxs = await get_question_field(ans_b64)

  re_qst = []
  for rbox in field_rboxs:
    score, qst_bboxs = rbox2bbox(rbox)
    if score > 0.2:
      img_re = cv2.polylines(img_re, [qst_bboxs], isClosed=True, color=(255, 255, 0), thickness=3)
      qst_field = four_point_transform(img2, qst_bboxs)
      re_qst.append(qst_field)

  # 3.4 bubble detection and OCR
  res = {}

  async def detector(qst):
    sharpen = deblur_image(qst)
    sharpen_b64 = img_to_base64(Image.fromarray(sharpen))
    qst_b64 = img_to_base64(Image.fromarray(qst))

    id_qst, bb = await asyncio.gather(
      get_id(sharpen_b64),
      get_bubble_answer(qst_b64)
    )
    bb_bboxs, bb_labels, bb_scores = bb[0], bb[1], bb[2]
    rebbox, relabels, rescores = [], [], []
    for i in range(0, len(bb_labels)):
      if bb_scores[i] > 0.7:
        rebbox.append(bb_bboxs[i])
        relabels.append(bb_labels[i])
        rescores.append(bb_scores[i])

    df = pd.DataFrame({
      "rebbox": rebbox,
      "relabels": relabels,
      "rescores": rescores
    })
    df_sorted = df.sort_values(by=["rebbox"])

    # for a -> d
    if (id_qst) and (id_qst[0].isnumeric()) and (int(id_qst[0]) > 0) and (int(id_qst[0]) < 41):
      re_format = ['A', 'B', 'C', 'D', 'E']

      grading[id_qst[0]] = df_sorted.relabels.values

      re = df_sorted.relabels.values.tolist()
      res[id_qst[0]] = [re_format[i] for i in range(len(re)) if re[i] == 1]

      logging.debug(f"RAW {id_qst[0]} - {re}")
      logging.info(f"PROCESS {res}")

  task_list = []
  for qst in re_qst:
    task_list.append(asyncio.create_task(detector(qst)))
  for task in task_list:
    await task

  return res


class OMRGradingServiceServicer(vision_pb2_grpc.OMRGradingServiceServicer):
  def OMRGrading(self, request, context):
    # Get images
    bboxs_re = []
    try:
      b64 = request.b64_img

      # processing
      grading_res_dict = asyncio.run(grading_omr(b64))
      grading_res = json.dumps(grading_res_dict, indent=2)

      return vision_pb2.OMRGradingResponse(id=request.id, result=str(grading_res),
                                           status=vision_pb2.OmrServiceStatus.Value("SUCCESS"))

    except Exception as e:
      logging.error(f"{SERVICE_NAME} ERROR. Reason: {e} - trace: {traceback.format_exc()}")
      return vision_pb2.OMRGradingResponse(id=request.id, result=str(bboxs_re),
                                           status=vision_pb2.OmrServiceStatus.Value('ERROR_IN_GRADING'))
