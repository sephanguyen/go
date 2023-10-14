import copy

import pandas as pd

from script.convert_bestco_format.config import SUBJECT_MAPPING_CONVERT


def get_student_subject(register_df):
  # get subject of student from register slot

  tmp_register_df = register_df[["student_id", "grade",
                                 "literature_slot", "math_slot", "en_slot", "science_slot", "social_science_slot",
                                 "sd_literature_slot", "sd_math_slot", "sd_en_slot", "sd_science_slot",
                                 "sd_social_slot"]]
  tmp_register = pd.DataFrame(columns=["student_id", "grade", "subject"])

  for i in range(len(tmp_register_df)):
    subject = []
    if (tmp_register_df["math_slot"][i] + tmp_register_df["sd_math_slot"][i]) > 0:
      subject.append("math")
    if (tmp_register_df["en_slot"][i] + tmp_register_df["sd_en_slot"][i]) > 0:
      subject.append("english")
    if (tmp_register_df["science_slot"][i] + tmp_register_df["sd_science_slot"][i]) > 0:
      subject.append("science")
    if (tmp_register_df["social_science_slot"][i] + tmp_register_df["sd_social_slot"][i]) > 0:
      subject.append("social_science")
    if (tmp_register_df["literature_slot"][i] + tmp_register_df["sd_literature_slot"][i]) > 0:
      subject.append("literature")
    subject = list(set(subject))
    dict_re = {
      "student_id": tmp_register_df["student_id"][i],
      "grade": tmp_register_df["grade"][i],
      "subject": subject
    }

    tmp_register = pd.concat([tmp_register, pd.DataFrame([dict_re])], ignore_index=True)

  tmp_register = tmp_register.explode("subject").dropna()
  return tmp_register


def compare(tmp_register, student_subject_fmt):
  tmp_register_1 = copy.deepcopy(tmp_register)
  student_subject_fmt_1 = copy.deepcopy(student_subject_fmt)

  student_subject_fmt_1["me"] = [1] * len(student_subject_fmt_1)
  tmp_register_1["me"] = [0] * len(tmp_register_1)

  left_re = pd.merge(student_subject_fmt_1, tmp_register_1, how="left", left_on=["student_id", "grade", "subject"],
                     right_on=["student_id", "grade", "subject"])
  assert len(left_re[left_re["me_y"].isnull()].reset_index()) == 0

  right_re = pd.merge(student_subject_fmt_1, tmp_register_1, how="right", left_on=["student_id", "grade", "subject"],
                      right_on=["student_id", "grade", "subject"])
  assert (len(right_re[right_re["me_x"].isnull()])) == 0

  inner_re = pd.merge(student_subject_fmt_1, tmp_register_1, how="inner", left_on=["student_id", "grade", "subject"],
                      right_on=["student_id", "grade", "subject"])
  assert len(inner_re[inner_re["me_y"].isnull()]) == 0

  print("DONE")


def get_avail_time(student_time, center_time):
  raw_stu_day = pd.merge(student_time, center_time, how="cross").query(
    "(date_x == date_y) & (time_period_x == time_period_y) & (available_or_not == 1) & (open_or_not == 1)").drop_duplicates()
  raw_stu_day = raw_stu_day[["student_id", "date_x", "time_period_x", "available_or_not", "time_period_y", "date_y",
                             "open_or_not"]].drop_duplicates()
  raw_stu_day.query("(date_x != date_y) | (time_period_x != time_period_y)")
  raw_stu_day.drop_duplicates()
  raw_stu_day = raw_stu_day.rename(columns={"date_x": "date", "time_period_x": "time_period"})
  return raw_stu_day[["student_id", "date", "time_period", "open_or_not"]]


def compare_day(raw, fmt):
  fmt["me1"] = [1] * len(fmt)
  raw["me0"] = [0] * len(raw)

  tmp1 = pd.merge(fmt, raw, how="left", left_on=["date", "time_period", "student_id"],
                  right_on=["date", "time_period", "student_id"]).drop_duplicates()
  assert len(tmp1.query("me0.isnull()")) == 0

  tmp2 = pd.merge(fmt, raw, how="right", left_on=["date", "time_period", "student_id"],
                  right_on=["date", "time_period", "student_id"]).drop_duplicates()
  assert len(tmp2.query("me1.isnull()")) == 0
  print("DONE")


def get_raw_stu_info(register_df, student_df, center_day):
  raw_student_subject = get_student_subject(register_df)
  raw_student_time = get_avail_time(student_df, center_day)
  raw_student_info = pd.merge(raw_student_subject, raw_student_time, how="cross").query(
    "(student_id_x == student_id_y)").drop_duplicates().reset_index()
  return raw_student_info.rename(columns={"student_id_x": "student_id"}).drop(columns=["student_id_y", "index"])


def compare_student_info(raw, fmt):
  raw["me0"] = [0] * len(raw)
  fmt["me1"] = [1] * len(fmt)

  tmp1 = pd.merge(fmt, raw, how="left", left_on=["date", "time_period", "student_id", "subject", "grade"],
                  right_on=["date", "time_period", "student_id", "subject", "grade"]).drop_duplicates()
  assert len(tmp1.query("me0.isnull()")) == 0

  tmp2 = pd.merge(fmt, raw, how="right", left_on=["date", "time_period", "student_id", "subject", "grade"],
                  right_on=["date", "time_period", "student_id", "subject", "grade"]).drop_duplicates()
  assert len(tmp2.query("me1.isnull()")) == 0
  print("DONE")


def get_teacher_subject_grade(teacher_subject_df):
  """
  This function get teacher and subject grading of teacher.
  """

  # SUBJECT_MAPPING = ["", "literature", "math", "english", "science", "social_science", "science", "social_science",
  #                    "",
  #                    "", "", "english", "", "literature", "math", "english"]
  tmp_teacher_subject_df = teacher_subject_df[["teacher_id", "grade_div", "subject_id", "available_or_not"]]
  tmp_teacher_subject_df["subject"] = tmp_teacher_subject_df.apply(lambda row: SUBJECT_MAPPING_CONVERT[row["subject_id"]],
                                                                   axis=1)
  tmp_teacher_subject_df = tmp_teacher_subject_df[
    (tmp_teacher_subject_df["subject"] != "") & (tmp_teacher_subject_df["available_or_not"] == 1)]

  grade = []
  for i in range(len(tmp_teacher_subject_df)):
    if (tmp_teacher_subject_df["grade_div"].reset_index()["grade_div"][i] == 1):
      grade.append([1, 2, 3, 4, 5])
    if (tmp_teacher_subject_df["grade_div"].reset_index()["grade_div"][i] == 2):
      grade.append([6, 7, 8, 9])
    if (tmp_teacher_subject_df["grade_div"].reset_index()["grade_div"][i] == 3):
      grade.append([10, 11, 12])
  tmp_teacher_subject_df["grade"] = grade
  tmp_teacher_subject_df = tmp_teacher_subject_df.explode("grade")
  tmp_teacher_subject_df = tmp_teacher_subject_df[["teacher_id", "subject", "grade"]]

  return tmp_teacher_subject_df.drop_duplicates()


def compare_teacher_subject(raw, fmt):
  raw["me0"] = [0] * len(raw)
  fmt["me1"] = [1] * len(fmt)

  tmp1 = pd.merge(fmt, raw, how="left", left_on=["teacher_id", "subject_name", "grade"],
                  right_on=["teacher_id", "subject", "grade"]).drop_duplicates()
  assert len(tmp1.query("me0.isnull()")) == 0

  tmp2 = pd.merge(fmt, raw, how="right", left_on=["teacher_id", "subject_name", "grade"],
                  right_on=["teacher_id", "subject", "grade"]).drop_duplicates()
  assert len(tmp2.query("me1.isnull()")) == 0

  print("DONE")


def get_raw_teacher_avail_day(teacher_df, center_day_df):
  tmp_teacher_df = teacher_df[["teacher_id", "date", "time_period", "available_or_not"]]
  tmp_teacher_df = tmp_teacher_df[tmp_teacher_df["available_or_not"] == 1]

  tmp = pd.merge(tmp_teacher_df, center_day_df, how="cross").query(
    "(date_x == date_y) & (time_period_x == time_period_y) & open_or_not == 1")
  tmp = tmp[["teacher_id", "date_x", "time_period_x"]].rename(
    columns={"date_x": "date", "time_period_x": "time_period"})
  return tmp


def compare_avail_day_teacher(raw, fmt):
  raw["me0"] = [0] * len(raw)
  fmt["me1"] = [1] * len(fmt)

  tmp1 = pd.merge(fmt, raw, how="left", left_on=["date_teacher", "time_period_teacher", "teacher_id"],
                  right_on=["date", "time_period", "teacher_id"]).drop_duplicates()
  assert len(tmp1.query("me0.isnull()")) == 0

  tmp2 = pd.merge(fmt, raw, how="right", left_on=["date_teacher", "time_period_teacher", "teacher_id"],
                  right_on=["date", "time_period", "teacher_id"]).drop_duplicates()
  assert len(tmp2.query("me1.isnull()")) == 0

  print("DONE")


def get_teacher_info(teacher_subject_df, teacher_df, center_day_df):
  raw_teacher_subject = get_teacher_subject_grade(teacher_subject_df)
  raw_avail_day_teacher = get_raw_teacher_avail_day(teacher_df, center_day_df)

  tmp = pd.merge(raw_teacher_subject, raw_avail_day_teacher, how="cross").query("(teacher_id_x == teacher_id_y)")

  return tmp.rename(columns={"teacher_id_x": "teacher_id"})[
    ["teacher_id", "subject", "grade", "date", "time_period"]].drop_duplicates()


def compare_avail_day_teacher(raw, fmt):
  raw["me0"] = [0] * len(raw)
  fmt["me1"] = [1] * len(fmt)

  tmp1 = pd.merge(fmt, raw, how="left",
                  left_on=["teacher_id", "subject_name", "date_center", "time_period_center", "grade"],
                  right_on=["teacher_id", "subject", "date", "time_period", "grade"]).drop_duplicates()
  assert len(tmp1.query("me0.isnull()")) == 0

  tmp2 = pd.merge(fmt, raw, how="right",
                  left_on=["teacher_id", "subject_name", "date_center", "time_period_center", "grade"],
                  right_on=["teacher_id", "subject", "date", "time_period", "grade"]).drop_duplicates()
  assert len(tmp2.query("me1.isnull()")) == 0

  print("DONE")


def compare_potential_slot(student_info_fmt, fmt_teacher_info, student_course, teacher_course):
  potential_before_naming = pd.merge(student_info_fmt, fmt_teacher_info, how="inner",
                                     left_on=["grade", "time_period", "date", "subject"],
                                     right_on=["grade", "time_period_teacher", "date_teacher", "subject_name"])
  potential_after_naming = pd.merge(student_course, teacher_course, how="inner",
                                    left_on=["available_day", "available_time", "grade", "subject"],
                                    right_on=["available_day", "available_time", "grade", "subject"])
  assert len(potential_before_naming) == len(potential_after_naming)
  print("DONE")


def compare_name_list(std_name_list, fmt):
  std_name_list["me0"] = [0] * len(std_name_list)
  fmt["me1"] = [1] * len(fmt)

  tmp1 = pd.merge(std_name_list, fmt, how="left", left_on=["student_id", "student_idx"],
                  right_on=["student_id_applied", "student"]).drop_duplicates()
  assert len(tmp1.query("me0.isnull()")) == 0

  tmp2 = pd.merge(std_name_list, fmt, how="right", left_on=["student_id", "student_idx"],
                  right_on=["student_id_applied", "student"]).drop_duplicates()
  assert len(tmp2.query("me1.isnull()")) == 0

  print("DONE")


def compare_teacher_name_list(raw, fmt):
  raw["me0"] = [0] * len(raw)
  fmt["me1"] = [1] * len(fmt)

  tmp1 = pd.merge(raw, fmt, how="left", left_on=["teacher_id", "teacher_idx"],
                  right_on=["teacher_id", "teacher_idx"]).drop_duplicates()
  assert len(tmp1.query("me0.isnull()")) == 0

  tmp2 = pd.merge(raw, fmt, how="right", left_on=["teacher_id", "teacher_idx"],
                  right_on=["teacher_id", "teacher_idx"]).drop_duplicates()
  assert len(tmp2.query("me1.isnull()")) == 0

  print("DONE")


def compare_stu_info(raw, fmt):
  raw["me0"] = [0] * len(raw)
  fmt["me1"] = [1] * len(fmt)

  tmp1 = pd.merge(raw, fmt, how="left", left_on=["student_id", "grade", "subject", "date", "time_period"],
                  right_on=["student_id_applied", "grade_int", "subject", "date_center",
                            "time_period_center"]).drop_duplicates()
  assert len(tmp1.query("me0.isnull()")) == 0

  tmp2 = pd.merge(raw, fmt, how="right", left_on=["student_id", "grade", "subject", "date", "time_period"],
                  right_on=["student_id_applied", "grade_int", "subject", "date_center",
                            "time_period_center"]).drop_duplicates()
  assert len(tmp2.query("me1.isnull()")) == 0

  print("DONE")


def compare_day_list_to_idx(day_list, fmt):
  day_list["me0"] = [0] * len(day_list)
  fmt["me1"] = [1] * len(fmt)

  tmp1 = pd.merge(day_list, fmt, how="left", left_on=["date", "date_idx"],
                  right_on=["date_center", "date_idx"]).drop_duplicates()
  assert len(tmp1.query("me0.isnull()")) == 0

  tmp2 = pd.merge(day_list, fmt, how="right", left_on=["date", "date_idx"],
                  right_on=["date_center", "date_idx"]).drop_duplicates()
  assert len(tmp2.query("me1.isnull()")) == 0

  print("DONE")


def compare_avail_teacher_day(raw, fmt_teacher_day):
  raw["me0"] = [0] * len(raw)
  fmt_teacher_day["me1"] = [1] * len(fmt_teacher_day)

  tmp1 = pd.merge(fmt_teacher_day, raw, how="left",
                  left_on=["teacher_id", "date_center", "time_period_center", "date_idx"],
                  right_on=["teacher_id", "date_x", "time_period", "date_idx"]).drop_duplicates()
  assert len(tmp1.query("me0.isnull()")) == 0

  tmp2 = pd.merge(fmt_teacher_day, raw, how="right",
                  left_on=["teacher_id", "date_center", "time_period_center", "date_idx"],
                  right_on=["teacher_id", "date_x", "time_period", "date_idx"]).drop_duplicates()
  assert len(tmp2.query("me1.isnull()")) == 0

  print("DONE")


def compare_teacher_info(raw, fmt):
  raw["me0"] = [0] * len(raw)
  fmt["me1"] = [1] * len(fmt)

  tmp1 = pd.merge(fmt, raw, how="left", left_on=["teacher", "subject", "available_day", "available_time", "grade"],
                  right_on=["teacher_idx", "subject", "date_idx", "time_period", "grade_fmt"]).drop_duplicates()
  assert len(tmp1.query("me0.isnull()")) == 0

  tmp2 = pd.merge(fmt, raw, how="right", left_on=["teacher", "subject", "available_day", "available_time", "grade"],
                  right_on=["teacher_idx", "subject", "date_idx", "time_period", "grade_fmt"]).drop_duplicates()
  assert len(tmp2.query("me1.isnull()")) == 0

  print("DONE")