import grpc
import click

from concurrent import futures
from pkg.manabuf_py.aphelios.vision.v1 import vision_pb2_grpc

from internal.aphelios.modules.omr.python.omr.usecase import bubble_detector, answer_detector, id_question_detector, \
  question_field_detector, omr_grading
from internal.aphelios.modules.omr.python.omr.utils.configure_loader import loader


@click.command()
@click.option('--secret_config', required=True, help="Path of secrect config, which is decrypted")
@click.option('--config', required=True, help="Path of config file")
def serve(secret_config, config):
  loader.minio.set_configure_path(config)
  loader.minio.set_secrect_config_path(secret_config)

  server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))

  #1. Id question detector
  vision_pb2_grpc.add_IdQuestionDetectorServiceServicer_to_server(
    id_question_detector.IdQuestionDetectorServiceServicer(),
    server
  )

  #2. Answer detector
  vision_pb2_grpc.add_AnswerDetectorServiceServicer_to_server(
    answer_detector.AnswerDetectorServiceServicer(),
    server
  )

  #3. Question field detector
  vision_pb2_grpc.add_QuestionFieldDetectorServiceServicer_to_server(
    question_field_detector.QuestionFieldDetectorServiceServicer(),
    server
  )

  #4. Bubble detector
  vision_pb2_grpc.add_BubbleDetectorServiceServicer_to_server(
    bubble_detector.BubbleDetectorServiceServicer(),
    server
  )

  #5
  vision_pb2_grpc.add_OMRGradingServiceServicer_to_server(
    omr_grading.OMRGradingServiceServicer(), server
  )

  server.add_insecure_port('[::]:50051')
  server.start()
  server.wait_for_termination()


if __name__ == '__main__':
  serve()
