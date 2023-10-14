import logging
import os
import sys

current_path = os.getcwd()
sys.path += [f"{current_path}"]

import psycopg
import csv
import click
import traceback

from internal.scheduling.script.pull_data.config import dataraw_path, ip_host

log_format = "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
logging.basicConfig(format=log_format, level=logging.ERROR)

logger = logging.getLogger()


@click.command()
@click.option('--student_avail_time_id', prompt='available time of student id',
              help='available time of student')
@click.option('--teacher_avail_time_id', prompt='available time of teacher id',
              help='available time of teacher')
@click.option('--time_slot_master_id', prompt='timeslot sheet id:',
              help='timeslot sheet')
@click.option('--center_opening_slot', prompt='opening time at a center id:',
              help='opening time at a center')
@click.option('--applied_slot',
              prompt='Info about register slot of each student id:',
              help='Info about register slot of each student')
@click.option('--teacher_subject',
              prompt='Subject which each teacher can handle id:',
              help='Subject which each teacher can handle')
@click.option('--user',
              prompt='User Id login:',
              help='user login to cloudsql')
@click.option('--password',
              prompt='password:',
              help='Password for user')
def main(student_avail_time_id, teacher_avail_time_id, time_slot_master_id,
		center_opening_slot, applied_slot, teacher_subject, user, password):
	output_path = dataraw_path
	os.makedirs(output_path, exist_ok=True)

	conn_str = f"dbname=calendar user={user} password={password} port=5432 hostaddr={ip_host}"
	with psycopg.connect(conn_str) as conn:
		# Open a cursor to perform database operations
		with conn.cursor() as cur:
			try:

				# 1. get the student info
				cur.execute(
						f"""SELECT * FROM student_available_slot_master WHERE run_time_id=\'{student_avail_time_id}\' """
				)
				cur.fetchone()
				header = ["id", "year", "period", "center_num", "student_id", "date",
				          "time_period", "available_or_not", "resource_path", "created_at", "updated_at", "deleted_at", "run_time_id"]
				with open(f"{output_path}/student_available_slot_master.csv", "w",
				          newline="") as stu_f:
					writer = csv.writer(stu_f)
					writer.writerow(header)
					writer.writerows(cur)
				logger.info("Done - 1/5 student info")

				# 2. get the teacher info
				cur.execute(
						f"""SELECT * FROM teacher_available_slot_master WHERE run_time_id=\'{teacher_avail_time_id}\' """
				)
				cur.fetchone()
				header = ["id", "year",	"period",	"center_num", "teacher_id",	"date",	"time_period", "available_or_not", "resource_path", "created_at", "updated_at", "deleted_at", "run_time_id"]
				with open(f"{output_path}/teacher_available_slot_master.csv", "w",
				          newline="") as teacher_f:
					writer = csv.writer(teacher_f)
					writer.writerow(header)
					writer.writerows(cur)
				logger.info("Done - 2/5 teacher info")

				# 3. get the applied_slot
				cur.execute(
						f"""SELECT * FROM applied_slot WHERE run_time_id=\'{applied_slot}\' """
				)
				cur.fetchone()
				header = ["id", "year",	"period",	"center_num", "student_id", "enrollment_status",
				          "grade","student_name","applied_slot","literature_slot",
				          "math_slot",	"en_slot",	"science_slot",	"social_science_slot",
				          "other_slot_1", "other_slot_2", "other_slot_3", "other_slot_4",
				          "other_slot_5", "other_slot_6", "other_slot_7", "other_slot_8",
				          "other_slot_9", "other_slot_10","sd_literature_slot",
				          "sd_math_slot",	"sd_en_slot",	"sd_science_slot",
				          "sd_social_slot",	"sd_other_slot_1",	"sd_other_slot_2",
				          "sd_other_slot_3",	"sd_other_slot_4", "sd_other_slot_5",
				          "sd_other_slot_6",	"sd_other_slot_7",	"sd_other_slot_8",
				          "sd_other_slot_9",	"sd_other_slot_10",	"preferred_gender",
				          "sibling_should_be_same_time",
				          "resource_path", "created_at", "updated_at", "deleted_at", "run_time_id"
				          ]
				with open(f"{output_path}/applied_slot.csv", "w",
				          newline="") as f:
					writer = csv.writer(f)
					writer.writerow(header)
					writer.writerows(cur)
				logger.info("Done - 3/5 aplied slot info")

				# 4. get the center_opening_slot
				cur.execute(
						f"""SELECT * FROM center_opening_slot WHERE run_time_id=\'{center_opening_slot}\' """
				)
				cur.fetchone()
				header = ["id", "year",	"period",	"center_num", "date", "time_period",
				          "open_or_not", "available_or_not", "resource_path", "created_at",
				          "updated_at", "deleted_at", "run_time_id"]
				with open(f"{output_path}/center_opening_slot.csv", "w",
				          newline="") as f:
					writer = csv.writer(f)
					writer.writerow(header)
					writer.writerows(cur)
				logger.info("Done - 4/5 student info")

				# 5. get the teacher_subject
				cur.execute(
						f"""SELECT * FROM teacher_subject WHERE run_time_id=\'{teacher_subject}\' """
				)
				header = ["id", "teacher_id",	"grade_div", "subject_id",
				          "available_or_not", "resource_path", "created_at",
				          "updated_at", "deleted_at", "run_time_id"]
				cur.fetchone()
				with open(f"{output_path}/teacher_subject.csv", "w",
				          newline="") as f:
					writer = csv.writer(f)
					writer.writerow(header)
					writer.writerows(cur)
				logger.info("Done - 5/5 teacher - subject")

			except Exception:
				logger.info(traceback.print_exc(file=sys.stdout))

			# Make the changes to the database persistent
			conn.commit()


if __name__ == '__main__':
	main()
