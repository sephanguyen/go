import sys
import os

current_path = os.getcwd()
sys.path += [f"{current_path}",
             f"{current_path}/pkg/manabuf_py"]

import uvicorn
import jwt
import grpc
import traceback

from pkg.manabuf_py.scheduling.v1 import scheduling_pb2_grpc
from pkg.manabuf_py.scheduling.v1 import scheduling_pb2
from fastapi import FastAPI, Header, Response, status
from typing import Annotated, Union
from pydantic import BaseModel
from config import GRPC_PORT, HTTP_PORT

from fastapi.responses import FileResponse


class ReqScheduling(BaseModel):
	id_req: str
	name: str
	list_hard_constraints: list
	weight_soft_constraints: list
	student_avail_time_id: str
	teacher_avail_time_id: str
	time_slot_master_id: str
	center_opening_slot: str
	teacher_subject: str
	applied_slot: str


class ReqDownload(BaseModel):
	id_req: str
	scheduling_jobs_name: str


app = FastAPI()


def checkValidToken(token) -> bool:
	try:
		decoded = jwt.decode(token, options={"verify_signature": False})
		if (decoded["aud"] == "manabie-stag") & (decoded["iss"] == "manabie"):
			return True
	except jwt.InvalidTokenError:
		print('Token is invalid')
	return False


@app.get("/")
async def root(response: Response,
		authorization: Annotated[Union[str, None], Header()] = None):
	res = {"message": "Hello World",
	       "status": "Success"}
	token = authorization.replace("Bearer ", "")
	
	if checkValidToken(token):
		response.status_code = status.HTTP_200_OK
		return res
	response.status_code = status.HTTP_400_BAD_REQUEST
	return {"message": "invalid token"}


@app.get("/v1/download/", response_class=FileResponse)
async def downloadResult(scheduling_jobs_name: str, key: str
):
	key = key
	if checkValidToken(key):
		try:
			file_path = "./internal/scheduling/data/bestco/output/manabie_result_20230426.csv"
			return FileResponse(file_path, filename=f"result-scheduling-{scheduling_jobs_name}.csv")
		except:
			traceback.print_exc()
			jsonRes = {
				"scheduling_jobs_name": scheduling_jobs_name,
				"DownloadStatus": "FAILED, Have the scheduling jobs finished?"
			}
	
	return jsonRes


@app.post("/v1/scheduling/")
async def postScheduling(jsonReq: ReqScheduling,
		response: Response,
		authorization: Annotated[Union[str, None], Header()] = None
):
	token = authorization.replace("Bearer ", "")
	if checkValidToken(token):
		try:
			# create the request protobuff
			protoReq = scheduling_pb2.SchedulingRequest()
			
			protoReq.id_req = jsonReq.id_req
			protoReq.teacher_available_slot_master = jsonReq.teacher_avail_time_id
			protoReq.student_available_slot_master = jsonReq.student_avail_time_id
			protoReq.applied_slot = jsonReq.applied_slot
			protoReq.center_opening_slot = jsonReq.center_opening_slot
			protoReq.time_slot = jsonReq.time_slot_master_id
			protoReq.teacher_subject = jsonReq.teacher_subject
			protoReq.weight_soft_constraints.extend(jsonReq.weight_soft_constraints)
			protoReq.list_hard_constraints.extend(jsonReq.list_hard_constraints)
			
			print(protoReq)
			
			# setup chanel and send the message
			
			async with grpc.aio.insecure_channel(f'localhost:{GRPC_PORT}') as channel:
				stub = scheduling_pb2_grpc.SchedulingServiceStub(channel)
				res = await stub.Scheduling(protoReq)
			
			print(f"scheduling client received: {res}")
			
			jsonRes = {
				"id_res": res.id_res,
				"id_req": res.id_req,
				"SchedulingServiceStatus": "SUCCESS"
			}
			return jsonRes
		except:
			traceback.print_exc()
			jsonRes = {
				"id_res": 0,
				"id_req": 0,
				"SchedulingServiceStatus": "FAILED"
			}
			
			return jsonRes
	
	response.status_code = status.HTTP_400_BAD_REQUEST
	return {
		"message": "invalid token"
	}


if __name__ == "__main__":
	uvicorn.run("http_server:app", host="0.0.0.0", port=HTTP_PORT)
