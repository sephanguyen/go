import os
import sys

sys.path.append(os.path.join(os.path.dirname(os.path.abspath(__file__)), os.pardir, os.pardir, os.pardir))  # ./../../.. that mean ./backend
sys.path.append(os.path.join(os.path.dirname(os.path.abspath(__file__)), os.pardir, os.pardir, os.pardir,"./pkg/manabuf_py"))  # ./../../../pkg/manabuf_py that mean ./backend/pkg/manabuf_py

import grpc

from concurrent import futures
from config import GRPC_PORT

from pkg.manabuf_py.scheduling.v1 import scheduling_pb2_grpc
from internal.scheduling.modules.scheduling.usecase.create_scheduling_job import SchedulingServiceServicer


def serve():
  server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))

  scheduling_pb2_grpc.add_SchedulingServiceServicer_to_server(
    SchedulingServiceServicer(),
    server
  )

  server.add_insecure_port(f"[::]:{GRPC_PORT}")
  server.start()
  server.wait_for_termination()


if __name__ == '__main__':
  print("gRPC server starting...")
  serve()
