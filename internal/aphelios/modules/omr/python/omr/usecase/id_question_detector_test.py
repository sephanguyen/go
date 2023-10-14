import unittest

from unittest import mock
from pkg.manabuf_py.aphelios.vision.v1 import vision_pb2
from internal.aphelios.modules.omr.python.omr.usecase.id_question_detector import IdQuestionDetectorServiceServicer


class MockTest(unittest.TestCase):
  def setUp(self):
    value = vision_pb2.IdQuestionResponse(id="id1", ans_img_url="image_anszone.png", bin_img_url="img_bin.png",
                                          status=vision_pb2.OmrServiceStatus.Value("SUCCESS"))
    self.patcher = mock.patch(
      'modules.omr.python.omr.usecase.id_question_detector.IdQuestionDetectorServiceServicer.IdQuestionDetector',
      return_value=value)
    self.patcher.start()

  def test_IdQuestionDetector(self):
    expected = vision_pb2.IdQuestionResponse(id="id1", ans_img_url="image_anszone.png", bin_img_url="img_bin.png",
                                             status=vision_pb2.OmrServiceStatus.Value("SUCCESS"))

    id_question_detector = IdQuestionDetectorServiceServicer()
    val = id_question_detector.IdQuestionDetector(request=vision_pb2.IdQuestionRequest(id="id1", image_url="image.png"),
                                                  context="")
    self.assertEqual(val, expected, "value is not like expected!!!")

  def tearDown(self) -> None:
    self.patcher.stop()


if __name__ == '__main__':
  unittest.main()
