import unittest

from unittest import mock
from pkg.manabuf_py.aphelios.vision.v1 import vision_pb2
from internal.aphelios.modules.omr.python.omr.usecase.get_top_view import GetTopViewServiceServicer


class MockTest(unittest.TestCase):
  def setUp(self):
    value = vision_pb2.GetTopViewResponse(id="id1", image_url="image_topview.png",
                                          status=vision_pb2.OmrServiceStatus.Value("SUCCESS"))
    self.patcher = mock.patch('modules.omr.python.omr.usecase.get_top_view.GetTopViewServiceServicer.GetTopView',
                              return_value=value)
    self.patcher.start()

  def test_GetTopView(self):
    expected = vision_pb2.GetTopViewResponse(id="id1", image_url="image_topview.png",
                                             status=vision_pb2.OmrServiceStatus.Value("SUCCESS"))

    get_top_view = GetTopViewServiceServicer()
    val = get_top_view.GetTopView(request=vision_pb2.GetTopViewRequest(id="id1", image_url="image.png"), context="")
    self.assertEqual(val, expected, "value is not like expected!!!")

  def tearDown(self) -> None:
    self.patcher.stop()


if __name__ == '__main__':
  unittest.main()
