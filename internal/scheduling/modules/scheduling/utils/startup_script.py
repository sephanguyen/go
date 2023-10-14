from internal.scheduling.script.pull_data.config import data_formated_path, \
	dataraw_path


class StartupScript:
	script_ = ""
	
	def get_script(self):
		return self.script_
	
	# teacher_avail_time_id = request.teacher_available_slot_master
	# student_avail_time_id = request.student_available_slot_master
	#
	# applied_slot = request.applied_slot
	# center_opening_slot = request.center_opening_slot
	# time_slot_master_id = request.time_slot
	# weight_soft_constraints = request.weight_soft_constraints
	# list_hard_constraints = request.list_hard_constraints
	def set_script(self, teacher_avail_time_id, student_avail_time_id,
			applied_slot, center_opening_slot, time_slot_master_id, teacher_subject,
			user, password):
		self.script_ = f"""
#!/usr/bin/bash

# there is a folder /scheduling as internal/scheduling
cd /home/ubuntu
export PYTHONPATH=$PYTHONPATH:/home/ubuntu
mkdir -p {data_formated_path}
sudo chmod 777 /home/ubuntu/data

# 1. pull data
/home/ubuntu/miniconda3/bin/python3 ./internal/scheduling/script/pull_data/pull_data.py \
--student_avail_time_id={student_avail_time_id} \
--teacher_avail_time_id={teacher_avail_time_id} \
--time_slot_master_id={time_slot_master_id} \
--center_opening_slot={center_opening_slot} \
--applied_slot={applied_slot} \
--teacher_subject={teacher_subject} \
--user={user} \
--password={password}
2> /home/ubuntu/1_error_log_run_pull_data.txt >>/home/ubuntu/1_log_pull_data.txt


# 2. convert data to manabie format:
/home/ubuntu/miniconda3/bin/python3 ./internal/scheduling/script/convert_bestco_format/convert_format.py \
--student_csv_path={dataraw_path}/student_available_slot_master.csv \
--teacher_csv_path={dataraw_path}/teacher_available_slot_master.csv \
--teacher_subject_csv={dataraw_path}/center_opening_slot.csv \
--day_csv_path={dataraw_path}/center_opening_slot.csv \
--applied_slot={dataraw_path}/applied_slot.csv \
--output_folder={data_formated_path} \
2> /home/ubuntu/2_error_log_run_convert_data.txt >>/home/ubuntu/2_log_convert_data.txt



# 3. run scheduling job
/home/ubuntu/miniconda3/bin/python3 ./internal/scheduling/job/bestco/scheduling.py \
--teacher_csv_path={data_formated_path}/teacher_formated.csv \
--student_csv_path={data_formated_path}/student_course_formated.csv \
--result_path=/home/ubuntu/timetable_result.csv \
2> /home/ubuntu/3_error_log_run.txt >>/home/ubuntu/3_log_run.txt

# 3. waiting have result then shut down the VM
if [ -f "/home/ubuntu/timetable_result.csv" ]
then
sudo shutdown -h now
fi
"""
