import unittest

import pandas as pd

from script.convert_bestco_format.utils import get_teacher_info, get_raw_stu_info

student_path = "../../../data/bestco/input/raw/student_available_slot_master.csv"
teacher_path = "../../../data/bestco/input/raw/teacher_available_slot_master.csv"
teacher_subject = "../../../data/bestco/input/raw/teacher_subject.csv"
register = "../../../data/bestco/input/raw/applied_slot.csv"
available_day = "../../../data/bestco/input/raw/center_opening_slot.csv"

student_df = pd.read_csv(student_path)
teacher_df = pd.read_csv(teacher_path)
teacher_subject_df = pd.read_csv(teacher_subject)
register_df = pd.read_csv(register)
center_day = pd.read_csv(available_day)


class TestConvertFormat(unittest.TestCase):

  def test_number_potential_slot(self):
    global teacher_subject_df
    global teacher_df
    global register_df
    global student_df
    global center_day

    raw_teacher_info = get_teacher_info(teacher_subject_df, teacher_df, center_day)
    raw_student_info = get_raw_stu_info(register_df, student_df, center_day)

    student_course = pd.read_csv("../../../data/bestco/input/formated/student_course.csv")
    teacher_course = pd.read_csv("../../../data/bestco/input/formated/teacher_course.csv")

    potential_before_naming = pd.merge(raw_teacher_info, raw_student_info, how="inner",
                                       left_on=["grade", "time_period", "date", "subject"],
                                       right_on=["grade", "time_period", "date", "subject"])
    potential_after_naming = pd.merge(student_course, teacher_course, how="inner",
                                      left_on=["available_day", "available_time", "grade", "subject"],
                                      right_on=["available_day", "available_time", "grade", "subject"])
    assert len(potential_before_naming) == len(potential_after_naming)
    print("DONE")


if __name__ == '__main__':
  unittest.main()
