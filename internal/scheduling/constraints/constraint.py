import math

from config import NUM_STUDENT, NUM_TEACHER, NUM_LEVEL, NUM_DAY, NUM_SHIFT, \
	NUM_DAY_FIRST_HALF, NUM_SHIFT_FIRST_HALF, \
	NUM_SHIFT_SECOND_HALF, PERCENT_SLOT_AT_PRI, SUBJECT_MAPPING, CENTER_CAPACITY


class Constraint():
	"""
	This class include constraint list which be implemented.
	"""

	def __init__(self, constraint, model, list_of_constraints):
		self.all_students = range(NUM_STUDENT)
		self.all_teachers = range(NUM_TEACHER)
		self.all_subjects = range(len(constraint["merge_sub_list"]))
		self.all_level = range(NUM_LEVEL)
		self.all_day = range(NUM_DAY)
		self.all_shifts = range(NUM_SHIFT)
		self.constraint = constraint
		self.model = model

		self.num_constraints = 9
		self.list_of_constraints = list_of_constraints
		assert len(list_of_constraints) == self.num_constraints, \
			"the config is mismatch with number constraint." \
			" Please re-check LIST_HARD_CONSTRAINT values"

	def add_cnstr_teacher_only_teach_one_slot_at_a_time(self, teacher_slot):
		"""

		Returns:
		"""
		if self.list_of_constraints[5] == 0:
			return

		### in each time slot, teacher only teach 1 subject at a time.

		for t in self.all_teachers:
			for day in self.all_day:
				for shift in self.all_shifts:
					self.model.AddAtMostOne(teacher_slot[(t, sub, day, shift, level)]
					                        for sub in self.all_subjects
					                        for level in self.all_level
					                        )

	def add_cnstr_student_study_one_slot_at_a_time(self, classes):
		"""

		Returns:
		"""
		if self.list_of_constraints[2] == 0:
			return

		## Student is study at one class at a time
		for stu in self.all_students:
			for shift in self.all_shifts:
				for day in self.all_day:
					self.model.Add(sum([
						classes[(stu, t, sub, level, day, shift)]
						for t in self.all_teachers
						for level in self.all_level
						for sub in self.all_subjects
					]) <= 1)

	def add_cnstr_90_primary_time(self, primary_slot, classes):
		## distribute allocate student slot at primary time 90% timeslot
		### first half: 1st day to 11th day. (80% number of all day)
		### second half: 12nd day to 14rd day. (20% number of all day)
		### primary time = 85% of first half + 50% of second half

		## 90% slot of each student in primary time
		## primary time = 85% of first half + 50% of second half

		if self.list_of_constraints[0] == 0:
			return

		prime = 0
		for day in range(NUM_DAY_FIRST_HALF):
			for shift in range(NUM_SHIFT_FIRST_HALF):
				prime += 1
				primary_slot[(shift, day)] = 1

		for day in range(NUM_DAY_FIRST_HALF, NUM_DAY):
			for shift in range(NUM_SHIFT_SECOND_HALF):
				prime += 1
				primary_slot[(shift, day)] = 1

		for stu in self.all_students:
			self.model.Add(
					sum([
						classes[(stu, t, sub, level, day, shift)]
						for t in self.all_teachers
						for level in self.all_level
						for (shift, day) in primary_slot
						for sub in self.all_subjects
					]) <= int(PERCENT_SLOT_AT_PRI * self.constraint["stu_slot"][stu]))

	def add_cnstr_max_slot_per_day(self, classes):
		"""

		Returns:
		"""
		if self.list_of_constraints[1] == 0:
			return

		## make sure number slot < max slot per day
		for stu in self.all_students:
			for day in self.all_day:
				self.model.Add(sum([
					classes[(stu, t, sub, level, day, shift)]
					for t in self.all_teachers
					for sub in self.all_subjects
					for level in self.all_level
					for shift in self.all_shifts
				]) <= self.constraint["stu_slot_per_day"][stu])

	@staticmethod
	def max_workup_hours_for_staff(counted_shifts: any) -> bool:
		"""
		formula calculate staff work 8 hours + each 10 minutes break per day with
		hardcoded duration time slots is 1 hour, 10 minutes break time

		```
		counted_shifts * duration_shift + counted_shifts * break_time - break_time <= 8 hours

		<=>  counted_shifts * (duration_shift + break_time) - break_time <= 8 hours
		```

		note: `- break_time` is ignoring last break time

		:param counted_shifts: any
		:return: bool(is 8 workup hours for staff?)
		"""
		worked_hours = 8  # hours
		duration_shift = 60  # minutes
		break_time = 10  # minutes
		return counted_shifts * (
				duration_shift + break_time) - break_time <= worked_hours * 60

	def add_cnstr_max_shift_per_day_staff(self, teacher_shift, classes):
		"""
		Staff work up to 8 hours per day

		Returns:
		"""
		if self.list_of_constraints[7] == 0:
			return

		for t in self.all_teachers:
			for day in self.all_day:
				self.model.Add(
						self.max_workup_hours_for_staff(
								sum(teacher_shift[(t, day, shift)]
								    for shift in self.all_shifts)
						)
				)

		for t in self.all_teachers:
			for stu in self.all_students:
				for sub in self.all_subjects:
					for level in self.all_level:
						for day in self.all_day:
							for shift in self.all_shifts:
								self.model.Add(
										sum([teacher_shift[(t, day, shift)]]) < 1
								).OnlyEnforceIf(classes[(stu, t, sub, level, day, shift)].Not())

	def add_cnstr_merge_subject_with_the_same_ratio(self, classes, teacher_slot,
			student_subject_class):
		"""
		Due to teacher can teach multi subject, which have the same ratio at the same time.
		So we merge these subjects to one, in order to input to the scheduling system.

		Returns:
		"""
		# make sure number slot of subject < max slot per subject
		## Combine subject with the same ratio
		## All class of subject and merge subject under limit: math + eng_math + math_science <= total_slot(math)

		if (self.list_of_constraints[3] == 0 or self.list_of_constraints[4] == 0):
			return

		for stu in self.all_students:
			for sub in range(len(SUBJECT_MAPPING)):  # root subject
				"""
				- Merged subject is slot in which teacher can teach all subject in that merged subject.
				With merged subject English+Math is allocated for Teacher A on day 1 shift 0, then Teacher A can teach both English
				and Math in that slot. The condition for merge subject, all subject which have the same ratio can be merge to
				one slot, whose teacher can handle both of them

				- When make sure to have enough slot for each student for each subject. We count all subject slot and all
				merge subject is less than the number of subject slot registered.
				- In case, all subject in merged subject are registered subject by one student, then all that subject will overlap
				so we have to multiply to have enough slot for all of them.
				"""
				related_sub_list = []
				multiply_sub_list = []  # multiply list to make sure have enough slot merge subject is allocated for student.
				tmp_list = []
				stu_sub_raw = self.constraint[
					"stu_sub"]  # get map present registered subject of each student

				# create related subject list
				if (stu_sub_raw[stu][sub] == 1):
					for sub2 in self.all_subjects:
						if ((SUBJECT_MAPPING[sub] != self.constraint["merge_sub_list"][
							sub2]) & (
								SUBJECT_MAPPING[sub] in self.constraint["merge_sub_list"][
							sub2].split(
								"+"))):  # sub2 is a merge subject and sub in list merge_subject available
							count = [stu_sub_raw[stu][SUBJECT_MAPPING.index(sub)] for sub in
							         self.constraint["merge_sub_list"][sub2].split(
									         "+")]  # count in merge subject, does any subject which student register? 0 not, 1 is register [0,1,0]

							multiply_sub_list.append(
									sum(
											count))  # limit of slot for merge subject will be increase
							tmp_list.append(sum(
									count))  # for calculate lcm to get convert float to int. 1,5x + 2y < 3 -> 3x + 4y < 6
							related_sub_list.append(sub2)

						elif (
								SUBJECT_MAPPING[sub] == self.constraint["merge_sub_list"][
							sub2]):
							multiply_sub_list.append(1)
							related_sub_list.append(sub2)

					# lcm tmp
					upper_bound = int(math.prod(tmp_list))
					max_val = max(tmp_list) if len(tmp_list) > 0 else 0
					lcm = 0
					for i in range(int(max_val), upper_bound + 1):
						remainder_list = [i % int(tmp) if tmp > 0 else 0 for tmp in
						                  tmp_list]
						if sum(remainder_list) == 0:
							lcm = i

					### student don't study more than slot which they register
					self.model.Add(sum([
						(lcm // multiply_sub_list[related_sub_list.index(related_sub)] if
						 multiply_sub_list[related_sub_list.index(
								 related_sub)] != 0 else 0) *
						classes[(stu, t, related_sub, level, day, shift)]
						for t in self.all_teachers
						for level in self.all_level
						for day in self.all_day
						for shift in self.all_shifts
						for related_sub in related_sub_list  # merge subject
					]) <= int(lcm * self.constraint["slot_of_each_subject"][stu][sub]))

		#### in case teacher don't teach subject - sub at slot (day-shift) -> Slot at that time can't be true
		for day in self.all_day:
			for shift in self.all_shifts:
				for sub in self.all_subjects:
					for level in self.all_level:
						for t in self.all_teachers:
							self.model.Add(sum([
								classes[(stu, t, sub, level, day, shift)]
								for stu in self.all_students]) < 1
							               ).OnlyEnforceIf(
									teacher_slot[(t, sub, day, shift, level)].Not())

		## set value student subject class
		for t in self.all_teachers:
			for stu in self.all_students:
				for sub in self.all_subjects:
					for level in self.all_level:
						for day in self.all_day:
							for shift in self.all_shifts:
								self.model.Add(
										sum([student_subject_class[(stu, day, shift, sub)]]) < 1
								).OnlyEnforceIf(classes[(stu, t, sub, level, day, shift)].Not())

	def add_soft_cnstr_consecutive_subject_teacher(self, classes):

		### student not learn with a consecutive teacher and subject
		for t in self.all_teachers:
			for stu in self.all_students:
				for sub in self.all_subjects:
					for level in self.all_level:
						for day in self.all_day:
							for shift in range(NUM_SHIFT - 1):
								self.model.Add(
										(classes[(stu, t, sub, level, day, shift)]
										 +
										 classes[(stu, t, sub, level, day, shift + 1)]) <= 1
								)

	def add_cnstr_num_stu_with_ratio(self, classes):
		"""
		This constraint make sure number student in each class not greater than the limitation of subject.
		Args:
				classes:

		Returns:

		"""
		for t in self.all_teachers:
			for sub in self.all_subjects:
				for shift in self.all_shifts:
					for day in self.all_day:
						for level in self.all_level:
							self.model.Add(sum([
								classes[(stu, t, sub, level, day, shift)]
								for stu in self.all_students
							]) <= int(self.constraint["merge_ratio_list"][sub]))

	def add_soft_cnstr_consecutive_slot(self, student_slot, classes):
		"""
		This constraint makes students and teacher don't teach consecutive subjects with the same student    Returns:
		"""
		#############################
		## set value student slot base on class available
		for t in self.all_teachers:
			for stu in self.all_students:
				for sub in self.all_subjects:
					for level in self.all_level:
						for day in self.all_day:
							for shift in self.all_shifts:
								self.model.Add(
										sum([student_slot[(stu, day, shift)]]) < 1
								).OnlyEnforceIf(classes[(stu, t, sub, level, day, shift)].Not())

	def add_cnstr_center_capacity(self, classes):
		"""
		This constraint makes sure the total number of students learning at one time
		is less than the center capacity
		Returns:

		"""
		if (self.list_of_constraints[8] == 0):
			return

		for day in self.all_day:
			for shift in self.all_shifts:
				self.model.Add(
						sum([
							classes[(stu, t, sub, level, day, shift)]
							for stu in self.all_students
							for t in self.all_teachers
							for level in self.all_level
							for sub in self.all_subjects]
						) < int(CENTER_CAPACITY))

	def get_model(self):
		return self.model
