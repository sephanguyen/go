import os
import sys

current_path = os.getcwd()
sys.path += [f"{current_path}", f"{current_path}/internal/scheduling"]

import copy

import click
import pandas as pd

from utils import get_raw_stu_info, get_teacher_info


@click.command()
@click.option('--student_csv_path', prompt=True, default="./data/bestco/input/raw/student_available_slot_master.csv")
@click.option('--teacher_csv_path', prompt=True, default="./data/bestco/input/raw/teacher_available_slot_master.csv")
@click.option('--teacher_subject_csv', prompt=True, default="./data/bestco/input/raw/teacher_subject.csv")
@click.option('--center_opening_slot_path', prompt=True, default="./data/bestco/input/raw/center_opening_slot.csv")
@click.option('--applied_slot', prompt=True, default="./data/bestco/input/raw/applied_slot.csv")
@click.option('--output_folder', prompt=True, default="./data/bestco/input/raw")
def convert(student_csv_path, teacher_csv_path, teacher_subject_csv, center_opening_slot_path, applied_slot, output_folder):
  stu_list_df = student_index(student_csv=student_csv_path)
  teacher_list_df = teacher_index(teacher_csv=teacher_csv_path)
  teacher_subject_df = teacher_grading(teacher_csv=teacher_subject_csv)
  day_list_df = date_index(date_csv=center_opening_slot_path)

  student_df = convert_student(student_csv=student_csv_path, center_opening_slot_path=center_opening_slot_path, applied_slot=applied_slot,
                               stu_list_df=stu_list_df, day_list_df=day_list_df)

  teacher_df = convert_teacher(teacher_csv=teacher_csv_path, teacher_list_df=teacher_list_df,
                               teacher_subject_df=teacher_subject_df, day_list_df=day_list_df)

  student_df.to_csv(f"{output_folder}/student_course.csv")
  teacher_df.to_csv(f"{output_folder}/teacher_course.csv")

  print(f"NUM_STUDENT:" + str(len(student_df["student"].drop_duplicates())))
  print(f"NUM_STUDENT:" + str(len(teacher_df["teacher"].drop_duplicates())))
  print(f"NUM_DAY:" + str(len(day_list_df["date_idx"].drop_duplicates())))
  print(f"NUM_SHIFT:" + str(len(day_list_df["time_period"].drop_duplicates())))
  print("DONE")


def convert_student(student_csv, center_opening_slot_path, applied_slot, stu_list_df, day_list_df):
  """
  Convert subject format
  """
  student_df = pd.read_csv(student_csv)
  applied_df = pd.read_csv(applied_slot)
  applied_df = applied_df[
    ["year", "student_id", "period", "center_num", "grade", "math_slot", "en_slot", "social_science_slot",
     "science_slot", "literature_slot", "sd_math_slot", "sd_en_slot", "sd_social_slot", "sd_science_slot",
     "sd_literature_slot"]]

  applied_df["math_x"] = applied_df["math_slot"]
  applied_df["english_x"] = applied_df["en_slot"]
  applied_df["social_science_x"] = applied_df["social_science_slot"]
  applied_df["science_x"] = applied_df["science_slot"]
  applied_df["literature_x"] = applied_df["literature_slot"]

  applied_df["math_y"] = applied_df["sd_math_slot"]
  applied_df["english_y"] = applied_df["sd_en_slot"]
  applied_df["social_science_y"] = applied_df["sd_social_slot"]
  applied_df["science_y"] = applied_df["sd_science_slot"]
  applied_df["literature_y"] = applied_df["sd_literature_slot"]

  applied_df = applied_df.reset_index()
  ss_sub_list = []
  regular_sub_list = []
  for i in range(len(applied_df)):
    ss_list = []
    regular_list = []
    if applied_df["math_x"][i] > 0:
      ss_list.append("math")
    if applied_df["english_x"][i] > 0:
      ss_list.append("english")
    if applied_df["social_science_x"][i] > 0:
      ss_list.append("social_science")
    if applied_df["science_x"][i] > 0:
      ss_list.append("science")
    if applied_df["literature_x"][i] > 0:
      ss_list.append("literature")

    if applied_df["math_y"][i] > 0:
      regular_list.append("math")
    if applied_df["english_y"][i] > 0:
      regular_list.append("english")
    if applied_df["social_science_y"][i] > 0:
      regular_list.append("social_science")
    if applied_df["science_y"][i] > 0:
      regular_list.append("science")
    if applied_df["literature_y"][i] > 0:
      regular_list.append("literature")

    ss_sub_list.append(ss_list)
    regular_sub_list.append(regular_list)

  applied_df["season_subject"] = ss_sub_list
  applied_df["regular_subject"] = regular_sub_list

  """
  get available day of student
  """
  student_df = student_df[["year", "student_id", "period", "center_num", "date", "time_period", "available_or_not"]]
  student_applied_df = pd.merge(applied_df, student_df, how="cross", suffixes=("_applied", "_student")).query(
    "(student_id_applied == student_id_student) & (period_applied == period_student)")

  """
  get available day from center
  """
  day_df = pd.read_csv(center_opening_slot_path)
  day_df = day_df[["year", "period", "center_num", "date", "time_period", "open_or_not"]]
  std_applied_center_df = pd.merge(student_applied_df, day_df, how="cross", suffixes=("_std_applied", "_center")).query(
    "(date_std_applied == date_center) & (time_period_std_applied == time_period_center) "
    "& (available_or_not == 1) & (open_or_not == 1)")

  """
  get subject register of student
  """
  std_applied_center_df["subject"] = std_applied_center_df.apply(
    lambda row: list(set(row["season_subject"] + row["regular_subject"])), axis=1)

  # test 4
  tmp4 = copy.deepcopy(std_applied_center_df.explode("subject").dropna())

  ###################################################
  ###################################################
  """
  convert
  student id to student_idx
  date to date_idx
  """
  convert_stu = pd.merge(std_applied_center_df, stu_list_df, how="cross", suffixes=("std_app_center", "stu")).query(
    "(student_id_applied == student_id)"
  )

  convert_date = pd.merge(convert_stu, day_list_df[["date", "date_idx"]].drop_duplicates(), how="cross",
                          suffixes=("convert_stu", "day")).query(
    "(date_std_applied == date)"
  )

  """
  rename columns
  - student_idx -> student.
  - date idx -> available_day
  - period -> available_time
  """
  re_df = copy.deepcopy(convert_date)
  re_df["student"] = re_df["student_idx"]
  re_df["available_day"] = re_df["date_idx"]
  re_df["available_time"] = re_df["period"]
  re_df = re_df.explode("subject").reset_index()

  """
  filter out unused column
  """
  re_df["grade"] = re_df.apply(lambda row: "g" + str(row["grade"]), axis=1).to_list()
  re_df["location"] = re_df["center_num_applied"]
  re_df = re_df.dropna()

  ### test 5
  tmp5 = copy.deepcopy(re_df)

  #######################################################
  #######################################################

  fmt_student = tmp4
  stu_day = tmp5[["date_center", "available_day"]].drop_duplicates()

  applied_slot_df = pd.read_csv(applied_slot)
  center_day = pd.read_csv(center_opening_slot_path)
  day_list = date_index(date_csv=center_opening_slot_path)
  raw_student_info = get_raw_stu_info(applied_slot_df, student_df, center_day)

  ### convert student to student index
  stuid_to_idx = pd.merge(fmt_student, stu_list_df, how="cross").query("student_id_applied == student_id")
  stu_day_idx = pd.merge(stuid_to_idx, stu_day, how="cross").query("date_center_x == date_center_y")
  stu_day_idx["grade"] = stu_day_idx.apply(lambda row: "g" + str(row["grade"]), axis=1)

  st1 = pd.merge(raw_student_info, stu_list_df, how="cross").query("student_id_x == student_id_y")
  st2 = pd.merge(st1, day_list, how="cross").query("(date_x == date_y) & (time_period_x == time_period_y)")
  st3 = st2.rename(columns={"student_idx": "student", "date_idx": "available_day", "time_period_x": "available_time"})[
    ["student", "available_day", "available_time", "subject", "grade"]]
  st3["grade"] = st3.apply(lambda row: "g" + str(row["grade"]), axis=1)

  student_info_official = pd.merge(st3, stu_day_idx, how="inner",
                                   left_on=["student", "available_day", "available_time", "subject", "grade"],
                                   right_on=["student_idx", "available_day", "time_period_center", "subject", "grade"])

  student_info_official = student_info_official.drop(columns=["index", "student_id_applied",
                                                              "period_applied", "date_center_x",
                                                              "open_or_not", "center_num_applied",
                                                              "year_applied", "center_num", "date_center_y",
                                                              "year_student", "student_id_student",
                                                              "time_period_std_applied", "available_or_not", "year",
                                                              "period",
                                                              "center_num", "date_center_x", 'open_or_not',
                                                              'date_center_y'])

  student_info_official["prefer_teacher"] = ["t_0"] * len(student_info_official)
  student_info_official["number_absense"] = [0] * len(student_info_official)
  student_info_official["subject_absense"] = ["t_0"] * len(student_info_official)
  student_info_official["location"] = [77] * len(student_info_official)

  student_info_official["total_slot_x"] = student_info_official.apply(
    lambda row: row["math_x"] + row["english_x"] + row["social_science_x"] + row["science_x"] + row["literature_x"],
    axis=1)
  student_info_official["total_slot_y"] = student_info_official.apply(
    lambda row: row["math_y"] + row["english_y"] + row["social_science_y"] + row["science_y"] + row["literature_y"],
    axis=1)

  return student_info_official


def convert_teacher(teacher_csv, teacher_list_df, teacher_subject_df, day_list_df):
  """
  This function converts teacher info into Manabie format
  Args:
    teacher_csv:
    teacher_list_df:
    teacher_subject_df:
    day_list_df:

  Returns:
    teacher_info_df:

  """
  teacher_df = pd.read_csv(teacher_csv)
  raw_teacher_info = get_teacher_info(teacher_subject_df, teacher_df, day_list_df)

  # Filter teacher available day with center open day.
  teacher_avail_day = pd.merge(raw_teacher_info, day_list_df, how="cross").query(
    "(date_x == date_y) & (time_period_x == time_period_y)")

  # Convert teacher_id to teacher index
  teacher_to_idx = pd.merge(teacher_avail_day, teacher_list_df, how="cross").query("teacher_id_x == teacher_id_y")
  teacher_to_idx["grade_fmt"] = teacher_to_idx.apply(lambda row: "g" + str(row["grade"]), axis=1)

  # Finalize the columns name of teacher info
  teacher_info_df = teacher_to_idx.rename(columns={"teacher_idx": "teacher", "date_idx": "available_day", "time_period_x": "available_time"})[
    ["teacher", "available_time", "available_day", "subject", "grade"]].drop_duplicates()
  teacher_info_df["priority"] = [0] * len(teacher_info_df)
  teacher_info_df["grade"] = teacher_info_df.apply(lambda row: "g" + str(row["grade"]), axis=1)
  teacher_info_df["location"] = [77] * len(teacher_info_df)

  return teacher_info_df.drop_duplicates()


def date_index(date_csv):
  date_df = pd.read_csv(date_csv)
  date_df = date_df[["date", "time_period", "open_or_not"]]
  date_df = date_df.drop_duplicates().sort_values(by="date")

  date_list_df = pd.DataFrame(date_df["date"].drop_duplicates().sort_values().to_list(),
                              columns=["date"])
  date_list_df["date_idx"] = [i for i in range(len(date_list_df))]

  date_idx = []
  for i in range(len(date_df)):
    tmp = date_list_df.index[date_list_df["date"] == date_df["date"][i]].to_list()
    date_idx.append(tmp[0])

  date_df["date_idx"] = date_idx

  return date_df


def teacher_grading(teacher_csv):
  teacher_df = pd.read_csv(teacher_csv)

  # 1: primary school
  # 2: middle school
  # 3: high school
  grade_list = [
    [1, 2, 3, 4, 5],
    [6, 7, 8, 9],
    [10, 11, 12]
  ]

  teacher_df["grade"] = [grade_list[int(teacher_df["grade_div"][i]) - 1] for i in range(len(teacher_df))]
  return teacher_df


def student_index(student_csv):
  stu_df = pd.read_csv(student_csv)
  stu_list_df = pd.DataFrame(stu_df["student_id"].drop_duplicates().sort_values().to_list(),
                             columns=["student_id"])
  stu_idx = [f"st_{i + 1}" for i in range(len(stu_list_df))]
  stu_list_df["student_idx"] = stu_idx

  return stu_list_df


def teacher_index(teacher_csv):
  teacher_df = pd.read_csv(teacher_csv)
  teacher_list_df = pd.DataFrame(teacher_df["teacher_id"].drop_duplicates().sort_values().to_list(),
                                 columns=["teacher_id"])

  teacher_idx = [f"t_{i + 1}" for i in range(len(teacher_list_df))]
  teacher_list_df["teacher_idx"] = teacher_idx
  return teacher_list_df


if __name__ == '__main__':
  convert()
