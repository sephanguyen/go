import logging
import uuid
import yaml

from yaml.loader import SafeLoader
from internal.scheduling.modules.scheduling.utils.startup_script import \
	StartupScript
from pkg.manabuf_py.scheduling.v1 import scheduling_pb2, scheduling_pb2_grpc
from internal.scheduling.modules.scheduling.utils.compute_engine import \
	get_image_from_family, create_from_custom_image


class SchedulingServiceServicer(scheduling_pb2_grpc.SchedulingServiceServicer):
	
	def Scheduling(self, request, context):
		id_req = request.id_req
		
		id_res = str(uuid.uuid4())
		
		teacher_avail_time_id = request.teacher_available_slot_master
		student_avail_time_id = request.student_available_slot_master
		
		applied_slot = request.applied_slot
		center_opening_slot = request.center_opening_slot
		time_slot_master_id = request.time_slot
		teacher_subject = request.teacher_subject
		weight_soft_constraints = request.weight_soft_constraints
		list_hard_constraints = request.list_hard_constraints
		""" Steps 1 to 3 is the equivalent of the ./manifestfiles/reverse_string.yaml """
		
		print(teacher_subject)
		
		# get login
		with open('/decrypt/scheduling.secrets.decrypt.config.yaml', 'r') as f:
			data = yaml.load(f, Loader=SafeLoader)

		# Compute  instance
		image = get_image_from_family('staging-manabie-online', 'scheduling-env')
		print(image)

		startup_script = StartupScript()
		startup_script.set_script(teacher_avail_time_id=teacher_avail_time_id,
		                          student_avail_time_id=student_avail_time_id,
		                          applied_slot=applied_slot,
		                          center_opening_slot=center_opening_slot,
		                          time_slot_master_id=time_slot_master_id,
		                          teacher_subject=teacher_subject,
		                          user=data["postgresql"]["user"],
		                          password=data["postgresql"]["password"]
		                          )

		re = create_from_custom_image(project_id='staging-manabie-online',
		                              zone='asia-southeast1-b',
		                              network_link="global/networks/staging",
		                              instance_name=f'scheduling-instance-run-{id_req}',
		                              custom_image_link='projects/staging-manabie-online/global/images/scheduling-base-image-01',
		                              startup_script=startup_script
		                              )
		
		logging.info(re)
		return scheduling_pb2.SchedulingResponse(id_res=id_res, id_req=id_req,
		                                         status=scheduling_pb2.SchedulingServiceStatus.SUCCESS)
