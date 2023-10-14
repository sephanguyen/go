import copy
import sys

sys.path.append("./internal/scheduling/job/bestco")

import pandas as pd
import typing

from config import NUM_STUDENT, NUM_SUBJECT, NUM_SHIFT, NUM_TEACHER, NUM_DAY, \
  SUBJECT_MAPPING, SUBJECT_RATIO, \
  MAX, NUM_LEVEL, LEVEL_TIME


def grade_to_level(grade_idx):
  level_idx = 0
  if (grade_idx <= 5) and (grade_idx >= 1):  # elementary
    level_idx = 0
  elif (grade_idx >= 6) and (grade_idx <= 9):  # middle school
    level_idx = 1
  elif (grade_idx >= 10) and (grade_idx <= 11):  # high_school
    level_idx = 2
  elif (grade_idx == 12):  # grade 12
    level_idx = 3

  return level_idx


def allocate_subject_slot(result_df, slot_of_each_subject):
  """
  This function is post_process, allocate acctual subject of merge subject slot for each student.
  Subject which have less remain slot is more priority.
  Args:
    result_df:
    slot_of_each_subject:

  Returns:

  """
  slot_sub_list = copy.deepcopy(slot_of_each_subject)
  sub_re = [" " for _ in range(len(result_df))]

  # priority minus the single slot first
  for i in range(len(result_df)):
    student_idx = int(result_df["student"][i].split("_")[1]) - 1
    subject = result_df["subject"][i]
    if (len(subject.split("+")) < 2):  # single subject
      subject_idx = SUBJECT_MAPPING.index(subject)
      sub_re[i] = subject
      slot_sub_list[student_idx][subject_idx] = slot_sub_list[student_idx][subject_idx] - 1

  for i in range(len(result_df)):
    subject = result_df["subject"][i]
    subject_list = subject.split("+")

    if (len(subject_list) > 1):  # merged subject
      student_idx = int(result_df["student"][i].split("_")[1]) - 1
      slot_min = MAX
      subject_min = ""
      for sub in subject_list:
        if (slot_sub_list[student_idx][SUBJECT_MAPPING.index(sub)] < slot_min) & (
            slot_sub_list[student_idx][SUBJECT_MAPPING.index(sub)] > 0):
          subject_min = sub
          slot_min = slot_sub_list[student_idx][SUBJECT_MAPPING.index(sub)]
      if subject_min != "":
        sub_re[i] = subject_min
        slot_sub_list[student_idx][SUBJECT_MAPPING.index(subject_min)] = slot_sub_list[student_idx][
                                                                           SUBJECT_MAPPING.index(subject_min)] - 1
      else:
        sub_re[i] = "invalid"

  result_df["actual_subject"] = sub_re
  return result_df


def get_prefer_score(std_path):
  """
  This util present prefer score between student and teacher

    - stu_teach[student_idx][teacher_idx] = 0: student_idx don't be prefer to teach by teacher_idx
    - stu_teach[student_idx][teacher_idx] = 1: student_idx want be teached by teacher_idx
  Args:
    std_path: path to student csv
  Returns:
    stu_teach: List<List<int>>
  """
  student_df = pd.read_csv(std_path)
  stu_teach = []
  df = student_df[["student", "prefer_teacher"]].drop_duplicates().reset_index()[["student", "prefer_teacher"]]

  for _ in range(NUM_STUDENT):
    tmp = []
    for _ in range(NUM_TEACHER):
      tmp.append(0)
    stu_teach.append(tmp)

  if len(df["prefer_teacher"].dropna()) > 0:
    for i in range(len(df)):
      stu_idx = int(df["student"][i].split("_")[1]) - 1
      t_idx = int(df["prefer_teacher"][i].split("_")[1]) - 1

      stu_teach[stu_idx][t_idx] = 1

  return stu_teach

def add_merged_subject_for_student(student_csv, merge_subject_list):
  stu_df = pd.read_csv(student_csv)
  re_stu_df = copy.deepcopy(stu_df)

  for i in range(NUM_STUDENT):
    if len(stu_df[stu_df["student"] == f"st_{i+1}"]) > 0:
      t = stu_df[["student", "grade", "subject", "available_day", "available_time", "prefer_teacher",
                  "location"]].drop_duplicates().groupby(by="student").get_group(f"st_{i + 1}")
      subject_list = t["subject"].drop_duplicates().to_list()
      grade_list = t["grade"].drop_duplicates().to_list()
      avail_day_list = t["available_day"].drop_duplicates().to_list()
      avail_time_list = t["available_time"].drop_duplicates().to_list()
      location_list = t["location"].drop_duplicates().to_list()
      prefer_teacher_list = t["prefer_teacher"].drop_duplicates().to_list()
      new_sub_list = []

      for subject in subject_list:
        new_sub_list = new_sub_list + [merge_sub for merge_sub in merge_subject_list if
                                       (subject in merge_sub.split("+")) & (len(merge_sub.split("+")) > 1)]

      # append new merge subject list to root df.
      tmp_dict = {
        "student": f"st_{i + 1}",
        "grade": grade_list,
        "subject": new_sub_list,
        "available_time": avail_time_list,
        "available_day": avail_day_list,
        "prefer_teacher": prefer_teacher_list,
        "location": location_list
      }

      tmp_df = pd.DataFrame.from_dict([tmp_dict])
      tmp_df = tmp_df.explode(
        column="grade"
      ).explode(
        column="subject"
      ).explode(
        column="available_time"
      ).explode(
        column="available_day"
      ).explode(
        column="prefer_teacher"
      ).explode(
        column="location"
      )
      re_stu_df = pd.concat([re_stu_df, tmp_df])

  return re_stu_df


def add_merged_subject_for_teacher(teacher_csv, teacher_sub):
  teacher_df = pd.read_csv(teacher_csv)
  re_teacher_df = copy.deepcopy(teacher_df)

  for i in range(len(teacher_sub)):
    if len(teacher_sub[i]) > 0:
      # get day and time
      t = teacher_df[
        ["teacher", "available_time", "available_day", "location", "priority"]
      ].drop_duplicates(
      ).groupby(
        by="teacher"
      ).get_group(
        f"t_{i + 1}")
      day_list = t["available_day"].drop_duplicates().to_list()
      shift_list = t["available_time"].drop_duplicates().to_list()
      location_list = t["location"].drop_duplicates().to_list()
      priority_list = t["priority"].drop_duplicates().to_list()

      merge_sub_list = []
      merge_sub_grade_list = []

      # get grade each subject
      t2 = teacher_df[["teacher", "subject", "grade"]].drop_duplicates().groupby(by="teacher").get_group(f"t_{i + 1}")
      sub_list = t2["subject"].to_list()
      grade_list = t2["grade"].to_list()

      for set_merge_sub in teacher_sub[i]:
        tmp_grade = []

        for sub in set_merge_sub.split("+"):
          list_grade_sub = [grade_list[idx] for idx, sub_list_elem in enumerate(sub_list) if sub_list_elem == sub]
          tmp_grade = list(set(tmp_grade + list_grade_sub))

        merge_sub_grade_list += list(set(tmp_grade))
        merge_sub_list += [set_merge_sub]

      df_re = pd.DataFrame(
        columns=["teacher", "available_day", "available_time", "location", "priority", "subject", "grade"])
      tmp_dict = {
        "teacher": f"t_{i + 1}",
        "available_day": day_list,
        "available_time": shift_list,
        "location": location_list,
        "priority": priority_list,
        "subject": merge_sub_list,
        "grade": merge_sub_grade_list
      }
      df_re_2 = pd.DataFrame.from_dict([tmp_dict])

      df_tmp = pd.concat([df_re, df_re_2], ignore_index=True)
      df_tmp = df_tmp.explode(
        column="available_day").explode(
        column="available_time").explode(
        column="location").explode(
        column="priority").explode(
        column="grade").explode(
        column="subject")

      re_teacher_df = pd.concat([re_teacher_df, df_tmp], ignore_index=True)

  return re_teacher_df


def get_merge_subject(teacher_csv):
  """
  This function create and add merged subject from subjects which teacher can handle
  Args:
    teacher_csv:

  Returns:
      merge_subject_mapping_list <list<>>: list of subject, included new merged subject
      merge_subject_ratio_list <list<int>>: ratio of each subject in merge_subject_mapping_list 's order
  """
  merge_subject_mapping_list = copy.deepcopy(SUBJECT_MAPPING)
  merge_subject_ratio_list = copy.deepcopy(SUBJECT_RATIO)
  teacher_sub = [[] for _ in range(NUM_TEACHER)]
  teacher_df = pd.read_csv(teacher_csv)

  for t in teacher_df["teacher"].drop_duplicates():
    teacher_idx = int(t.split("_")[1]) - 1
    subject_list = teacher_df[["teacher", "subject"]].drop_duplicates().groupby(by="teacher").get_group(t)["subject"] \
      .to_list()
    ratio_list = [SUBJECT_RATIO[int(SUBJECT_MAPPING.index(sub))] for sub in subject_list]

    for ratio in ratio_list:
      idx = [i for i, x in enumerate(SUBJECT_RATIO) if x == ratio]
      if (len(idx) > 1):
        tmp_sub = [SUBJECT_MAPPING[i] for i in idx if SUBJECT_MAPPING[i] in subject_list]
        tmp_sub = list(set(tmp_sub))
        tmp_sub.sort()
        merge_sub = "+".join([str(elem) for elem in tmp_sub])

        if (merge_sub not in merge_subject_mapping_list) & (len(tmp_sub) > 1):
          merge_subject_mapping_list.append(merge_sub)
          merge_subject_ratio_list.append(ratio)

          teacher_sub[teacher_idx].append(merge_sub)

  return merge_subject_mapping_list, merge_subject_ratio_list, teacher_sub


def get_all_stu_sub_slot(stu_path):
  """
    Get all slots of subject for each student
    list_student_subject[1][2] = 3 => student index 1 have 3 slot of subject index 2
  Args:
    stu_path:

  Returns:
    list_student_subject: list<list<int>>
  """
  df = pd.read_csv(stu_path)
  df["math_x"] = df["math_x"].fillna(0)
  df["english_x"] = df["english_x"].fillna(0)
  df["literature_x"] = df["literature_x"].fillna(0)
  df["science_x"] = df["science_x"].fillna(0)
  df["social_science_x"] = df["social_science_x"].fillna(0)
  df["subject_absense"] = df["subject_absense"].fillna("")

  list_student_subject = []
  for _ in range(NUM_STUDENT):
    subject_list = []
    for sub in range(NUM_SUBJECT):
      subject_list.append(0)
    list_student_subject.append(subject_list)

  for i in range(len(df)):
    student_idx = int(df["student"][i].split("_")[1]) - 1
    num_absense = df["number_absense"][i]
    sub_absense = df["subject_absense"][i]

    # SUBJECT_MAPPING = ["math", "english", "literature", "science", "social_science"]
    list_student_subject[student_idx][0] = df["math_x"][i] + df["math_y"][i]
    list_student_subject[student_idx][1] = df["english_x"][i] + df["english_y"][i]
    list_student_subject[student_idx][2] = df["literature_x"][i] + df["literature_y"][i]
    list_student_subject[student_idx][3] = df["science_x"][i] + df["science_y"][i]
    list_student_subject[student_idx][4] = df["social_science_x"][i] + df["social_science_y"][i]

    if sub_absense in SUBJECT_MAPPING:
      list_student_subject[student_idx][SUBJECT_MAPPING.index(sub_absense)] = list_student_subject[student_idx][
                                                                                SUBJECT_MAPPING.index(
                                                                                  sub_absense)] + num_absense

  return list_student_subject


def get_all_stu_slot(stu_path):
  """
    return all slot of each student in period time.
    list_stu_slot[stu_idx] =3 => the student_idx have 3 slot in that time.
  Args:
    stu_path:
  Returns:
    list_stu_slot : list<int>
  """
  df = pd.read_csv(stu_path)
  df["number_absense"] = df["number_absense"].fillna(0)
  df["total_slot_x"] = df["total_slot_x"].fillna(0)
  df["total_slot_y"] = df["total_slot_y"].fillna(0)

  list_stu_slot = []

  for _ in range(NUM_STUDENT):
    list_stu_slot.append(0)

  for i in range(len(df)):
    stu_idx = int(df["student"][i].split("_")[1]) - 1
    num_absense = int(df["number_absense"][i])
    total_regular_slot = int(df["total_slot_x"][i])
    total_seasonal_slot = int(df["total_slot_y"][i])

    list_stu_slot[stu_idx] = num_absense + total_regular_slot + total_seasonal_slot

  return list_stu_slot


def get_stu_slot_per_day(stu_day, all_stu_slot):
  """
    return list max slot per day of each student.
    list_stu_slot_per_day[student_idx] = 3 => student_idx take maximize 3 slot per day.
  Args:
    stu_day:
    all_stu_slot:
  Returns:
    list_stu_slot_per_day: list<int>
  """
  list_stu_slot_per_day = []

  for i in range(len(all_stu_slot)):
    if sum(stu_day[i]) > 0:
      list_stu_slot_per_day.append(1 + (all_stu_slot[i] // sum(stu_day[i])))
    else:
      list_stu_slot_per_day.append(0)

  return list_stu_slot_per_day


def get_stu_sub(stu_df, subject_list):
  """
    return list subject that student would like to learn.
    list_stu_sub[student_idx][subject_idx] = 1 => student index would like to learn subject_idx.
    list_stu_sub[student_idx][subject_idx] = 0 => student index would not like to learn subject_idx.
  Args:
    stu_path:
  Returns:
    list_stu_sub: list<list<int>>
  """
  df = stu_df.reset_index()

  list_stu_sub = []
  d = []
  for _ in range(NUM_STUDENT):
    sub_list = []
    for _ in range(len(subject_list)):
      sub_list.append(0)
    list_stu_sub.append(copy.deepcopy(sub_list))
    d.append(copy.deepcopy(sub_list))

  for i in range(len(df)):
    stu_idx = int(df["student"][i].split("_")[1]) - 1
    for k in range(len(subject_list)):
      if df["subject"][i] == subject_list[k]:
        list_stu_sub[stu_idx][k] = 1

  return list_stu_sub


def get_stu_level(stu_path):
  """
  Return list contain grading of each student.
  list_stu_grade[student_idx][level] = 1 => student_index is at level.
  list_stu_grade[student_idx][level] = 1 => student_index is not at level.
  :param stu_path:
  :return: list_stu_grade: list<list<int>>
  """
  df = pd.read_csv(stu_path)

  list_stu_grade = []

  for _ in range(NUM_STUDENT):
    list_grade = []
    for _ in range(NUM_LEVEL):
      list_grade.append(0)
    list_stu_grade.append(list_grade)

  for i in range(len(df)):
    stu_idx = int(df["student"][i].split("_")[1]) - 1
    grade_idx = int(df["grade"][i].replace("g", ""))

    level_idx = grade_to_level(grade_idx=grade_idx)

    list_stu_grade[stu_idx][level_idx] = 1

  return list_stu_grade

def get_stu_avail_time(student_path):
  """
  return list available day and shift of each student based on the register time
  and their grade
  Args:
    student_path:

  Returns:
    list_stu_time[student_idx][day_idx][shift_idx] = 0 => this day and shift when student can't study
    list_stu_time[student_idx][day_idx][shift_idx] = 1 => this day and shift when student can study
  """
  student_df = pd.read_csv(student_path)

  list_stu_time = []

  for _ in range(NUM_STUDENT):
    list_day = []
    for _ in range(NUM_DAY):
      list_shift = []
      for _ in range(NUM_SHIFT):
        list_shift.append(0)
      list_day.append(list_shift)
    list_stu_time.append(list_day)
    
  for i in range(len(student_df)):
    stu_idx = int(student_df["student"][i].split("_")[1]) - 1
    day_idx = student_df["available_day"][i]
    shift_idx = student_df["available_time"][i]
    grade = int(student_df["grade"][i].replace("g", ""))
    level = grade_to_level(grade)
    
    if (day_idx < NUM_DAY) & (str(shift_idx) in LEVEL_TIME[level]):
      list_stu_time[stu_idx][day_idx][shift_idx] = 1

  return list_stu_time

def get_stu_avail_day(stu_path):
  """
  return list of available day of each student.
  list_stu_day[student_idx][day_idx]
  :param stu_path:
  :return:
  """
  df = pd.read_csv(stu_path)
  list_stu_day = []

  for _ in range(NUM_STUDENT):
    list_day = []
    for _ in range(NUM_DAY):
      list_day.append(0)
    list_stu_day.append(list_day)

  for i in range(len(df)):
    if df["available_day"][i] < NUM_DAY:
      stu_idx = int(df["student"][i].split("_")[1]) - 1
      list_stu_day[stu_idx][df["available_day"][i]] = 1

  return list_stu_day


def get_stu_avail_shift(stu_path):
  """
  return list of available shift of student.
  list_stu_shift[student_idx][shift_idx]
  :param stu_path:
  :return:
  """
  df = pd.read_csv(stu_path)
  list_stu_shift = []

  for _ in range(NUM_STUDENT):
    list_shift = []
    for _ in range(NUM_SHIFT):
      list_shift.append(0)
    list_stu_shift.append(list_shift)

  for i in range(len(df)):
    stu_idx = int(df["student"][i].split("_")[1]) - 1
    list_stu_shift[stu_idx][df["available_time"][i]] = 1  # start shift is 0

  return list_stu_shift


def get_stu_pre_teacher(stu_path):
  """
  return list of prefer teacher of each student
  list_stu_teacher[student_idx][teacher_idx]
  :param stu_path:
  :return:
  """
  df = pd.read_csv(stu_path)
  list_stu_teacher = []

  for _ in range(NUM_STUDENT):
    list_teacher = []
    for _ in range(NUM_TEACHER):
      list_teacher.append(0)
    list_stu_teacher.append(list_teacher)

  if len(df["prefer_teacher"].dropna()) > 0:
    for i in range(len(df)):
      stu_idx = int(df["student"][i].split("_")[1]) - 1
      teacher_idx = int(df["prefer_teacher"][i].split("_")[1]) - 1
      list_stu_teacher[stu_idx][teacher_idx] = 1

  return list_stu_teacher


def get_teacher_sub_level(teacher_df, subject_list):
  '''
  return list of subject and grade which teacher can handle. value in array (0,1) 0 teacher can't handle subject at grade
  list_teacher_sub[teacher_idx][subject_idx][level_idx] = 1 => teacher idx can study subject_idx at level_idx
  list_teacher_sub[teacher_idx][subject_idx][level_idx] = 0 => teacher idx can NOT study subject_idx at level_idx
  Args:
    teacher_csv:
    subject_list:
    ratio_list:

  Returns:

  '''
  df = teacher_df

  list_teacher_sub = []

  for _ in range(NUM_TEACHER):
    list_sub = []
    for _ in range(len(subject_list)):
      list_grade = []
      for _ in range(NUM_LEVEL):
        list_grade.append(0)
      list_sub.append(list_grade)
    list_teacher_sub.append(list_sub)

  for i in range(len(df)):
    teacher_idx = int(df["teacher"][i].split("_")[1]) - 1
    grade_idx = int(df["grade"][i].replace("g", ""))

    level_idx = grade_to_level(grade_idx=grade_idx)

    for k in range(len(subject_list)):
      if df["subject"][i] == subject_list[k]:
        list_teacher_sub[teacher_idx][k][level_idx] = 1

  return list_teacher_sub


def get_teacher_avail_day(teacher_csv):
  """
  return list of available day of teacher.
  list_teacher_day[teacher_idx][day_idx]
  :param teacher_csv:
  :return:
  """
  df = pd.read_csv(teacher_csv)

  list_teacher_day = []

  for _ in range(NUM_TEACHER):
    list_day = []
    for _ in range(NUM_DAY):
      list_day.append(0)
    list_teacher_day.append(list_day)

  for i in range(len(df)):
    if df["available_day"][i] < NUM_DAY:
      teacher_idx = int(df["teacher"][i].split("_")[1]) - 1
      list_teacher_day[teacher_idx][df["available_day"][i]] = 1

  return list_teacher_day


def get_teacher_avail_shift(teacher_csv):
  """
  return list of available time shift of teacher.
  list_teacher_shift[teacher_idx][shift_idx] = 1 => the teacher_idx can teach at shift_idx and vice versa.
  :param teacher_csv:
  :return:
  """
  df = pd.read_csv(teacher_csv)
  list_teacher_shift = []

  for _ in range(NUM_TEACHER):
    list_shift = []
    for _ in range(NUM_SHIFT):
      list_shift.append(0)
    list_teacher_shift.append(list_shift)

  for i in range(len(df)):
    teacher_idx = int(df["teacher"][i].split("_")[1]) - 1
    list_teacher_shift[teacher_idx][df["available_time"][i]] = 1

  return list_teacher_shift

def get_teacher_available_time(teacher_path):
  """
  This function get availble date_time of each teacher
  Args:
    teacher_df:

  Returns:
    teacher_avail_time[teacher_idx][date_idx][time_idx] = 0 -> that time is not available
    teacher_avail_time[teacher_idx][date_idx][time_idx] = 1 -> that time is available

  """
  teacher_df = pd.read_csv(teacher_path)

  teacher_avail_time = []

  for _ in range(NUM_TEACHER):
    list_teacher = []
    for _ in range(NUM_DAY):
      list_day = []
      for _ in range(NUM_SHIFT):
        list_day.append(0)
      list_teacher.append(list_day)
    teacher_avail_time.append(list_teacher)

  for i in range(len(teacher_df)):
    teacher_idx = int(teacher_df["teacher"][i].split("_")[1]) - 1
    day_idx = teacher_df["available_day"][i]
    shift_idx = teacher_df["available_time"][i]
    if day_idx < NUM_DAY:
      teacher_avail_time[teacher_idx][day_idx][shift_idx] = 1

  return teacher_avail_time

def get_merge_list():
  pass


def get_all_stu_slot_types(stu_path) -> typing.Dict[int, typing.Dict[str, typing.Dict[str, int]]]:
  """
    return all slot types(seasonal, regular, absense) of each student in period time.
  Args:
    stu_path: str
  Returns:
    list_stu_slot_types : {
      1: { # -> student id
        'math': {
          "seasonal_slot": 3,
          "sd_slot": 2,
        },
        'english': {
          "seasonal_slot": 5,
          "sd_slot": 4,
        },
      },
      ...
    }

  """

  df = pd.read_csv(stu_path)
  df["number_absense"] = df["number_absense"].fillna(0)
  df["subject_absense"] = df["subject_absense"].fillna(0)
  df["total_slot_x"] = df["total_slot_x"].fillna(0)
  df["total_slot_y"] = df["total_slot_y"].fillna(0)
  df["subject"] = df["subject"].fillna(0)

  df["math_x"] = df["math_x"].fillna(0)
  df["english_x"] = df["english_x"].fillna(0)
  df["literature_x"] = df["literature_x"].fillna(0)
  df["science_x"] = df["science_x"].fillna(0)
  df["social_science_x"] = df["social_science_x"].fillna(0)

  df["math_y"] = df["math_y"].fillna(0)
  df["english_y"] = df["english_y"].fillna(0)
  df["literature_y"] = df["literature_y"].fillna(0)
  df["science_y"] = df["science_y"].fillna(0)
  df["social_science_y"] = df["social_science_y"].fillna(0)

  list_stu_slot_types = {}

  for i in range(len(df)):
    stu_idx = int(df["student"][i].split("_")[1]) - 1
    subject_absense = df["subject_absense"][i] if df["subject_absense"][i] in SUBJECT_MAPPING else ""
    total_absense_slot = int(df["number_absense"][i])

    math_x = int(df["math_x"][i])
    english_x = int(df["english_x"][i])
    literature_x = int(df["literature_x"][i])
    science_x = int(df["science_x"][i])
    social_science_x = int(df["social_science_x"][i])

    math_y = int(df["math_y"][i])
    english_y = int(df["english_y"][i])
    literature_y = int(df["literature_y"][i])
    science_y = int(df["science_y"][i])
    social_science_y = int(df["social_science_y"][i])

    list_stu_slot_types[stu_idx] = {
      'math': {
        "seasonal_slot": math_x,
        "sd_slot": math_y,
      },
      'english': {
        "seasonal_slot": english_x,
        "sd_slot": english_y,
      },
      'literature': {
        "seasonal_slot": literature_x,
        "sd_slot": literature_y,
      },
      'social_science': {
        "seasonal_slot": social_science_x,
        "sd_slot": social_science_y,
      },
      'science': {
        "seasonal_slot": science_x,
        "sd_slot": science_y,
      },
    }
    if total_absense_slot != 0:
      list_stu_slot_types[stu_idx][subject_absense]['sd_slot'] += total_absense_slot

  return list_stu_slot_types


def get_slot_type_of_a_student(list_stu_slot_types: typing.Dict[int, typing.Dict[str, typing.Dict[str, int]]],
                               student_idx: int, subject: str) -> str:
  """
    return type of slot based on student and subject in order
  Args:
    list_stu_slot_types: result of get_all_stu_slot_types
    student_idx: int
    subject: str
  Returns:
    slot type: str
  """

  # types: [seasonal_slot sd_slot]
  types = list(list_stu_slot_types[student_idx][subject].keys())
  index = 0

  while list_stu_slot_types[student_idx][subject][types[index]] == 0:
    index += 1
    if index == len(types):
      return ""

  list_stu_slot_types[student_idx][subject][types[index]] -= 1
  return types[index]


def post_process_assign_slot_types(student_csv_path: str, result_csv_path: str) -> None:
  df = pd.read_csv(result_csv_path)
  # get types(seasonal, SD) for all students
  stu_slot_types = get_all_stu_slot_types(stu_path=student_csv_path)

  # filter trash column name
  columns = [column for column in list(df) if 'Unnamed' not in column]
  resutl_df = pd.DataFrame(columns=columns)

  for index in range(len(df)):
    subject = df.loc[index, 'actual_subject']
    student_id = int(df.loc[index, 'student'].split('_')[1]) - 1

    data = {}
    slot_type = get_slot_type_of_a_student(stu_slot_types, student_id, subject)
    data["slot_type"] = slot_type

    for column in columns:
      data[column] = df.loc[index, column]

    tmp_df = pd.DataFrame([data])
    resutl_df = pd.concat([resutl_df, tmp_df], ignore_index=True)

  resutl_df.to_csv(result_csv_path)


def post_process_check_prefer_teacher(result_csv_path, stu_teacher_score):
  result_df = pd.read_csv(result_csv_path)

  re_prefer = []
  for i in range(len(result_df)):
    stu_idx = int(result_df["student"][i].split("_")[1]) - 1
    teacher_idx = int(result_df["teacher"][i].split("_")[1]) - 1

    re_prefer.append(stu_teacher_score[stu_idx][teacher_idx])

  result_df["is_prefer"] = re_prefer
  result_df.to_csv(result_csv_path)
