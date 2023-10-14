import sys

sys.path.append("./internal/scheduling/job/bestco")

from scheduling import scheduling, run_scheduling

import unittest
import click

import pandas as pd

from config import NUM_STUDENT, PERCENT_SLOT_AT_PRI, SUBJECT_MAPPING, \
	LEVEL_MAPPING, CENTER_CAPACITY, LEVEL_TIME
from utils import get_stu_sub, get_stu_avail_day, get_stu_avail_shift, \
	get_stu_pre_teacher, \
	get_teacher_avail_shift, get_teacher_avail_day, get_all_stu_slot, \
	get_stu_slot_per_day, \
	get_all_stu_sub_slot, get_merge_subject, add_merged_subject_for_teacher, \
	add_merged_subject_for_student, \
	get_all_stu_slot_types, \
	post_process_assign_slot_types, get_prefer_score, \
	post_process_check_prefer_teacher, grade_to_level, get_teacher_sub_level, \
	get_stu_level

from constraint import Constraint

student_csv_path = "./internal/scheduling/data/scheduling/input/student_course_formated.csv"
teacher_csv_path = "./internal/scheduling/data/scheduling/input/teacher_formated.csv"
result_csv_path = "./internal/scheduling/data/scheduling/output/result.csv"

avail_class_objective_value = 393312
total_remain_objective_value = 0


@click.command()
@click.option('--teacher_course_path', prompt="teacher course path:",
              default="./internal/scheduling/data/bestco/input/formated/teacher_course.csv",
              required=True)
@click.option('--student_course_path', prompt="student course path:",
              default="./internal/scheduling/data/bestco/input/formated/student_course.csv",
              required=True)
@click.option('--result_path', prompt="result path:",
              default="./internal/scheduling/data/bestco/output/manabie_run_final.csv",
              required=True)
@click.option('--is_run_scheduling', is_flag=True, default=False,
              prompt='Do you want to re-run scheduling?', required=True)
def run_test(teacher_course_path, student_course_path, result_path,
		is_run_scheduling):
	global stu_day
	global stu_slot
	global stu_slot_per_day
	global result_df
	
	global student_csv_path
	student_csv_path = student_course_path
	
	global teacher_csv_path
	teacher_csv_path = teacher_course_path
	
	global result_csv_path
	result_csv_path = result_path
	
	if is_run_scheduling:
		run_scheduling(teacher_csv_path=teacher_course_path,
		               student_csv_path=student_course_path,
		               result_path=result_path)
	
	print("Test...")
	unittest.main()


class TestAllocateLogic(unittest.TestCase):
	"""
	 This test, test allocate slot logic.
			+ 90% slot of each student in primary time
			+ make sure number slot < max slot per day
			+ number slot of each student register and applied. slot(math) + slot(math+eng) + ... + slot(math+<other_subject>) <= slot_register(math)
			+ number student in each slot <= capacity of that subject slot
	"""
	
	def test_max_slot_per_day(self):
		stu_slot = get_all_stu_slot(stu_path=student_csv_path)
		stu_day = get_stu_avail_day(stu_path=student_csv_path)
		stu_slot_per_day = get_stu_slot_per_day(stu_day, stu_slot)
		df = pd.read_csv(result_csv_path)
		
		allocate_by_day = df[["student", "day"]].value_counts()
		st_list = [st_idx[0] for st_idx in allocate_by_day.index.to_list()]
		
		for i in range(NUM_STUDENT):
			if f"st_{i + 1}" in st_list:
				num_slot = sum(allocate_by_day[f"st_{i + 1}"])
				num_day = len(allocate_by_day[f"st_{i + 1}"])
				self.assertLessEqual(num_slot / num_day, stu_slot_per_day[i],
				                     "Student be allocated more than max slot per day")
	
	def test_90percent_on_primary(self):
		stu_slot = get_all_stu_slot(stu_path=student_csv_path)
		df = pd.read_csv(result_csv_path)
		allocate_by_day = df[["student", "is_primary_slot"]].value_counts()
		st_list = [st_idx[0] for st_idx in allocate_by_day.index.to_list()]
		
		for i in range(NUM_STUDENT):
			num_slot = 0
			if f"st_{i + 1}" in st_list:
				if len(allocate_by_day[f"st_{i + 1}"]) > 1:
					num_slot = allocate_by_day[f"st_{i + 1}"][1]
				self.assertLessEqual(num_slot, int(PERCENT_SLOT_AT_PRI * stu_slot[i]),
				                     "Student be allocated more than 90% slot at primary time")
	
	def test_max_slot_per_student(self):
		"""
		This unit test make sure student don't study more than slot that they register
		note:
		 -  exception related to merge subject
				number_slot(eng) + number_slot(math) + number_slot(math+eng) <= number_slot(math) + number_slot(eng)
		Returns:
		"""
		merge_sub_list, merge_ratio_list, teacher_sub = get_merge_subject(
				teacher_csv_path)
		df = pd.read_csv(result_csv_path)
		slot_of_each_subject = get_all_stu_sub_slot(stu_path=student_csv_path)
		
		new_df = df[["student", "subject"]].groupby(
				by=["student", "subject"]).size().reset_index(name='time')
		st_list = new_df["student"].drop_duplicates().to_list()
		
		for i in range(NUM_STUDENT):
			if f"st_{i + 1}" in st_list:
				stu_sub_applied_list = [0 for _ in range(len(merge_sub_list))]
				tmp_df = new_df.groupby(by=["student"]).get_group(
						f"st_{i + 1}").reset_index()
				
				for sub_idx in range(len(tmp_df)):
					sub_name = tmp_df["subject"][sub_idx]
					count = tmp_df["time"][sub_idx]
					
					stu_sub_applied_list[merge_sub_list.index(sub_name)] = count
				
				# check
				for sub in SUBJECT_MAPPING:
					# get all single slot slot(math) + slot(math+eng) <= slot(math)
					right_side = slot_of_each_subject[i][SUBJECT_MAPPING.index(sub)]
					left_side = 0
					
					for sub_2 in merge_sub_list:
						if (sub_2 == sub) or (sub in sub_2.split("+")):
							left_side += stu_sub_applied_list[merge_sub_list.index(sub_2)]
					
					left_side = stu_sub_applied_list[merge_sub_list.index(sub)]
					
					self.assertLessEqual(left_side, right_side,
					                     f"ERROR: student st_{i + 1} is learn more than slot of {sub} which they register")
	
	def test_number_student_in_each_slot(self):
		"""
		This unit test make sure number student is not greater than capacity of each subject slot.
		time, day, sub, teacher => sum of all student <= ratio
		"""
		# merge subject
		merge_sub_list, merge_ratio_list, teacher_sub = get_merge_subject(
				teacher_csv_path)
		df = pd.read_csv(result_csv_path)
		
		df = df[["teacher", "subject", "day", "shift", "student"]].groupby(
				by=["teacher", "subject", "day", "shift"]).size().reset_index(
				name='number_student')
		
		for i in range(len(df)):
			subject = df["subject"][i]
			number_student_allocated = df["number_student"][i]
			capacity = merge_ratio_list[merge_sub_list.index(subject)]
			
			self.assertLessEqual(number_student_allocated, capacity,
			                     f"ERROR: number of student in class {subject} at slot {i} greater than the capacity")
	
	def test_teacher_only_handle_one_slot_at_time(self):
		"""
				This unit test make sure one teacher only teach at one slot at a time.
				number of subject in == 1.
				In case slot include 2 subject which have same ratio (eng, math). They will be presented by one merged subject eng+math
		"""
		df = pd.read_csv(result_csv_path)
		
		tmp = df[["teacher", "subject", "day", "shift"]].drop_duplicates().groupby(
				by=["teacher", "day", "shift"]).size().reset_index(name="num_subject")
		
		for i in range(len(tmp)):
			teacher = tmp["teacher"][i]
			day = tmp["day"][i]
			shift = tmp["shift"][i]
			self.assertLessEqual(tmp["num_subject"][i], 1,
			                     f"ERROR: Teacher {teacher} teach multi slot at {day} - shift {shift}")
	
	def test_stu_not_study_same_teacher_and_sub(self):
		"""
		This unit test make sure one student don't study with the same teacher or subject consecutively
		"""
		df = pd.read_csv(result_csv_path)
		tmp = df.groupby(by=["student", "day", "subject", "teacher"])
		
		for i in range(len(df)):
			value_tmp = tmp.get_group(
					(df["student"][i], df["day"][i], df["subject"][i], df["teacher"][i])
			).sort_values(by=['shift']).reset_index()
			
			# iterator all slot which have the same day, teacher, subject of a student.
			# if there are any 2 slot consecutive => conflict the constraints
			for k in range(len(value_tmp) - 1):
				if (len(value_tmp) > 1):
					student = value_tmp["student"][k]
					slot1 = value_tmp["shift"][k]
					slot2 = value_tmp["shift"][k + 1]
					self.assertGreater(value_tmp["shift"][k + 1] - value_tmp["shift"][k],
					                   1,
					                   f"ERROR: Student {student} study with the same teacher or subject consecutive"
					                   f"at slot {slot1} - {slot2} ")
	
	def test_allocate_seasonal_slot_first(self):
		"""
		This unit test make sure the seasonal slot must be allocated earlier than sd slots
		"""
		
		df = pd.read_csv(result_csv_path)
		df = df[["student", "actual_subject", "slot_type"]] \
			.sort_values(["student", "actual_subject"], ignore_index=False)
		
		"""
		check seasonal_slot must be earlier than sd_slot
		"""
		student_data = {}
		for index in range(len(df)):
			student_id = int(df.loc[index, 'student'].split('_')[1]) - 1
			subject = df.loc[index, 'actual_subject']
			slot_type = df.loc[index, 'slot_type']
			
			# set student id if not existed
			if student_data.get(student_id) == None:
				student_data[student_id] = {}
			
			# set subject if not existed
			if student_data[student_id].get(subject) == None:
				student_data[student_id][subject] = []
			
			student_data[student_id][subject].append(slot_type)
			
			if len(student_data[student_id][subject]) > 1:
				slot_types_of_subject = student_data[student_id][subject]
				previous_slot = slot_types_of_subject[-2]
				current_slot = slot_types_of_subject[-1]
				
				self.assertFalse(
						previous_slot == 'sd_slot' and current_slot == 'seasonal_slot',
						f'ERROR: seasonal_slot must be allocated earlier than sd_slot for {subject} of student {student_id + 1}'
				)
		
		"""
		double check: amount of defined slots are at least equal or greater than
									amount of allocated (seasonal & sd) slots
		"""
		stu_slot_types = get_all_stu_slot_types(stu_path=student_csv_path)
		for student_id in student_data:
			slot_types_of_subject = stu_slot_types[student_id]
			
			for subject in student_data[student_id]:
				allocated_seasonal_slots = [slot for slot in
				                            student_data[student_id][subject] if
				                            slot == 'seasonal_slot']
				allocated_sd_slots = [slot for slot in student_data[student_id][subject]
				                      if slot == 'sd_slot']
				
				self.assertFalse(
						slot_types_of_subject[subject]['seasonal_slot'] < len(
								allocated_seasonal_slots),
						f"""ERROR: wrong allocated slots, expected seasonal_slot >= allocated_seasonal_slots
                  but got amount seasonal_slots: {slot_types_of_subject[subject]['seasonal_slot']}'
                          amount allocated_seasonal_slots: {len(allocated_seasonal_slots)}'
              """
				)
				
				self.assertFalse(
						slot_types_of_subject[subject]['sd_slot'] < len(allocated_sd_slots),
						f"""ERROR: wrong allocated slots, expected sd_slot >= allocated_sd_slots
                  but got amount sd_slot: {slot_types_of_subject[subject]['sd_slot']}'
                          amount allocated_sd_slotssd_slot: {len(allocated_sd_slots)}'
              """
				)
	
	def test_prefer_teacher_slot(self):
		"""
		This unit test just check is there column prefer teacher in result
		and
		compare the objective value in 2 case: with or without soft constraint. With soft constraint, the objective values
		greater than without the soft constraints
		"""
		resutl_df = pd.read_csv(result_csv_path)
		
		self.assertTrue("is_prefer" in resutl_df.columns,
		                "ERROR: Missing column `is_prefer`. This allocation table may "
		                "missing allocate prefer teacher of each student")
		
		BASELINE_SCORE = 393311  # objective value without soft constraints
		
		self.assertGreaterEqual(avail_class_objective_value, BASELINE_SCORE,
		                        "ERROR: the objective score with soft constraints is less than baseline score")
	
	def test_teacher_teach_right_level_and_subject(self):
		"""
		This unit test checks if the teacher is allocated right to the class which
		has a subject and level that they can handle
		"""
		teacher_df = pd.read_csv(teacher_csv_path)
		result_df = pd.read_csv(result_csv_path)
		
		teacher_sub_level = get_teacher_sub_level(teacher_df=teacher_df,
		                                          subject_list=SUBJECT_MAPPING)
		
		for i in range(0, len(result_df)):
			re_level_idx = LEVEL_MAPPING.index(result_df["level"][i])
			re_sub_idx = SUBJECT_MAPPING.index(result_df["actual_subject"][i])
			re_teacher_idx = int(result_df["teacher"][i].split("_")[1]) - 1
			self.assertEqual(
					teacher_sub_level[re_teacher_idx][re_sub_idx][re_level_idx], 1,
					f"ERROR: Teacher_{re_teacher_idx + 1} dont teach {SUBJECT_MAPPING[re_sub_idx]} at level {LEVEL_MAPPING[re_level_idx]}")
	
	def test_split_merged_subject(self):
		"""
		- Check sum the number of splited subject is less or equal the total registered slot of student
		- The format of splited subject is splited completely
		"""
		
		df = pd.read_csv(result_csv_path)
		slot_of_each_subject = get_all_stu_sub_slot(stu_path=student_csv_path)
		
		self.assertTrue("actual_subject" in df.columns,
		                "ERROR: Missing column 'actual_subject' in the result file")
		for i in range(NUM_STUDENT):
			for k in range(len(SUBJECT_MAPPING)):
				tmp_df = df[(df["student"] == f"st_{i + 1}") & (
						df["actual_subject"] == SUBJECT_MAPPING[k])]
				self.assertLessEqual(len(tmp_df), slot_of_each_subject[i][k],
				                     "ERROR: Students study more than slot which they registered")
		
		for i in range(len(df)):
			self.assertFalse("+" in df["actual_subject"][i],
			                 "ERROR: actual_subject columns is not split completely")
	
	def test_student_same_level_in_class(self):
		"""
		This unit test check 2 things:
		- Result mapping correct from grade to level:
			[1,2,3,4,5] - elemetary; [6,7,8,9] - middle; [10,11] - high_school; [12] - 12_high

		- In each slot, that only include student in same level

		- Check teacher result mapping from grade to level is correct
		"""
		
		result_df = pd.read_csv(result_csv_path)
		student_df = pd.read_csv(student_csv_path)
		teacher_df = pd.read_csv(teacher_csv_path)
		
		# Result mapping correct from grade to level
		for i in range(len(result_df)):
			stu = result_df["student"][i]
			level_idx = LEVEL_MAPPING.index(result_df["level"][i])
			
			actual_grade = int(
					student_df[student_df["student"] == stu][
						["student", "grade"]].drop_duplicates().reset_index()["grade"][
						0].replace("g", ""))
			
			self.assertLessEqual(level_idx, grade_to_level(actual_grade),
			                     f"ERROR: mapping failed from grade {actual_grade} to level")
		
		# In each slot, that only include student in same level
		## slot = teacher, day, time
		tmp_df = result_df[["teacher", "day", "shift", "subject",
		                    "level"]].drop_duplicates().reset_index()
		for i in range(len(tmp_df)):
			teacher = tmp_df["teacher"][i]
			day = tmp_df["day"][i]
			shift = tmp_df["shift"][i]
			
			tmp = tmp_df.groupby(by=["teacher", "day", "shift"]).get_group(
					(teacher, day, shift)).reset_index()
			self.assertEqual(1, len(tmp),
			                 "ERROR: there are more than level in this slot")
		
		# In each slot teacher, subject and level. Check level is in list of level which teacher can handle for that subject
		teacher_df = teacher_df[
			["teacher", "subject", "grade"]].drop_duplicates().reset_index()
		for i in range(len(teacher_df)):
			grade_list = teacher_df.groupby(by=["teacher", "subject"]).get_group(
					(teacher_df["teacher"][i], teacher_df["subject"][i])).reset_index()[
				"grade"].to_list()
			teacher_level_idx_list = [grade_to_level(int(grade.replace("g", ""))) for
			                          grade in grade_list]
			teacher_idx = teacher_df["teacher"][i].split("_")[1]
			
			allocate_level_list = result_df[
				(result_df["actual_subject"] == teacher_df["subject"][i]) & (
						result_df["teacher"] == f"Teacher_{teacher_idx}")
				]["level"].to_list()
			allocate_level_idx_list = [LEVEL_MAPPING.index(level_name) for level_name
			                           in allocate_level_list]
			
			# print(f"{teacher_idx} - {allocate_level_idx_list} - {teacher_level_idx_list}")
			if len(allocate_level_list) > 0:
				for allocate_level_idx in allocate_level_idx_list:
					self.assertTrue(allocate_level_idx in teacher_level_idx_list,
					                "ERROR: teacher is located to different level which they can handle")
	
	def test_minimize_remain_slot(self):
		"""
		- This unit test check the objective value of function minimize the remain slot of each student
		- The objective value should be in range(0, full_slot)
			Note:
				full_slot: total slot registered of all student.
		"""
		slot_of_each_subject = get_all_stu_sub_slot(stu_path=student_csv_path)
		total_slot_register = sum(
				[sum(slot_of_each_subject[i]) for i in
				 range(len(slot_of_each_subject))])
		
		self.assertGreaterEqual(total_remain_objective_value, 0,
		                        "ERROR: seem total avail slot is greater than register slot")
		self.assertLessEqual(total_remain_objective_value, total_slot_register,
		                     "ERROR: seem there error of all slot register")
	
	def test_max_slot_per_staff(self):
		"""
		Test if the staff work no more than 8 hours per day.

		This function reads the result csv file and groups the shifts by teacher and day.
		It then asserts that the Constraint.max_workup_hours_for_staff function returns True
		for each teacher and day combination.

		:return: None
		"""
		df = pd.read_csv(result_csv_path)[["teacher", "day", "shift"]]
		results = df.groupby(["teacher", "day"])["shift"].count()
		
		for teacher_and_day, counted_shift in results.items():
			teacher, day = teacher_and_day
			self.assertTrue(
					Constraint.max_workup_hours_for_staff(counted_shift),
					f"ERROR: {teacher} works more than 8 hours on {day}",
			)
	
	def test_center_capacity(self):
		"""
		This test check the total number of students learning at the same time
		is less than center capacity or not
		Returns:
		"""
		
		df = pd.read_csv(result_csv_path)[["day", "shift", "student"]]
		results = df.groupby(by=["day", "shift"]).size().reset_index(
				name="number_student")
		
		for i in range(len(results)):
			day = results.iloc[i]["day"]
			shift = results.iloc[i]["shift"]
			self.assertLessEqual(
					results.iloc[i]["number_student"], CENTER_CAPACITY,
					f"ERROR: At day {day} - shift {shift}. The total student greater than center capacity"
			)
	
	def test_student_study_at_the_right_shift(self):
		"""
		This test check the avail shift of the student base on their grade.
		The config base on variable `LEVEL_TIME`
		Ex: elemetary student not study in the morning shift -> they don't available
		at the shift 0,1,2,3

		"""
		result_df = pd.read_csv(result_csv_path)[
			["day", "shift", "student", "level"]]
		
		for i in range(len(result_df)):
			stu_idx = int(result_df["student"][i].split("_")[1]) - 1
			shift = result_df["shift"][i]
			level_idx = LEVEL_MAPPING.index(result_df["level"][i])
			
			self.assertTrue(str(shift) in LEVEL_TIME[level_idx],
			                f"[ERROR] student {stu_idx + 1} "
			                f"at the level {LEVEL_MAPPING[level_idx]} "
			                f"cannot study at the shift_{shift}")


if __name__ == '__main__':
	run_test()
