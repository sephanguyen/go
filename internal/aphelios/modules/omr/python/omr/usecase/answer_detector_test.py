import unittest

from unittest import mock
from pkg.manabuf_py.aphelios.vision.v1 import vision_pb2
from internal.aphelios.modules.omr.python.omr.usecase.answer_detector import AnswerDetectorServiceServicer


class MockTest(unittest.TestCase):
  def setUp(self):
    value = vision_pb2.AnswerDetectorResponse(id="id1", image_url="image_result.png", bubble_choice=[],
                                              status=vision_pb2.OmrServiceStatus.Value("SUCCESS"))
    self.patcher = mock.patch(
      'modules.omr.python.omr.usecase.answer_detector.AnswerDetectorServiceServicer.AnswerDetector',
      return_value=value)
    self.patcher.start()

  def test_AnswerDetector(self):
    expected = vision_pb2.AnswerDetectorResponse(id="id1", image_url="image_result.png", bubble_choice=[],
                                                 status=vision_pb2.OmrServiceStatus.Value("SUCCESS"))

    answer_detector = AnswerDetectorServiceServicer()
    val = answer_detector.AnswerDetector(request=vision_pb2.AnswerDetectorRequest(id="id1"), context="")
    self.assertEqual(val, expected, "value is not like expected!!!")

  def tearDown(self) -> None:
    self.patcher.stop()


if __name__ == '__main__':
  unittest.main()
