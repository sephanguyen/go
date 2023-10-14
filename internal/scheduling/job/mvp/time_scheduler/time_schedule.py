import click
import pandas as pd
from ortools.sat.python import cp_model

NUM_STUDENT = 33
NUM_TEACHER = 7
NUM_SUBJECT = 3  # toan, ly, hoa
NUM_GRADE = 4  # 6,7,8,9
DAY_OF_WEEK = 6  # (t2->t7)
NUM_SHIFT = 4  # (14h-16h) , (16h-18h), (18-20), (20-22)

SUBJECT_MAPPING = ["Toan", "Ly", "Hoa"]
SHIFT_MAPPING = ["14h-16h", "16h-18h", "18h-20h", "20h-22h"]


# soft_constraint
# Subject
## student want learn some subject.
def get_student_subject_req(csv_path):
  student_df = pd.read_csv(csv_path)
  student_df[["Student", "subject"]]

  stu_sub_req = []
  for i in range(NUM_STUDENT):
    stu_sub_req.append([0, 0, 0])

  for i in range(len(student_df)):
    student_index = int(student_df["Student"][i].split("_")[1]) - 1

    if student_df["subject"][i].lower().strip() == "toan":
      stu_sub_req[student_index][0] = 1
    if student_df["subject"][i].lower().strip() == "ly":
      stu_sub_req[student_index][1] = 1
    if student_df["subject"][i].lower().strip() == "hoa":
      stu_sub_req[student_index][2] = 1

  return stu_sub_req


## teacher can handle some subject.
def get_teacher_subject_req(csv_path):
  teacher_df = pd.read_csv(csv_path)
  teacher_df[["Teacher", "subject"]]

  teacher_sub_req = []
  for i in range(NUM_TEACHER):
    teacher_sub_req.append([0, 0, 0])

  for i in range(len(teacher_df)):
    teacher_index = int(teacher_df["Teacher"][i].split("_")[1]) - 1

    if teacher_df["subject"][i].lower().strip() == "toan":
      teacher_sub_req[teacher_index][0] = 1
    if teacher_df["subject"][i].lower().strip() == "ly":
      teacher_sub_req[teacher_index][1] = 1
    if teacher_df["subject"][i].lower().strip() == "hoa":
      teacher_sub_req[teacher_index][2] = 1

  return teacher_sub_req


# Time
## student available time.
def get_std_time(csv_path):
  student_df = pd.read_csv(csv_path)

  std_time_slot = []

  for _ in range(NUM_STUDENT):
    time_slot = []
    for __ in range(DAY_OF_WEEK):
      time_slot.append([0, 0, 0, 0])  # 4 shift per day.

    std_time_slot.append(time_slot)

  for i in range(len(student_df)):
    student_index = int(student_df["Student"][i].split("_")[1]) - 1

    day = int(student_df["day_of_week"][i]) - 2  # convert from day of week to index day

    if student_df["time_available"][i].lower().strip() == "14h-16h":
      std_time_slot[student_index][day][0] = 1
    if student_df["time_available"][i].lower().strip() == "16h-18h":
      std_time_slot[student_index][day][1] = 1
    if student_df["time_available"][i].lower().strip() == "18h-20h":
      std_time_slot[student_index][day][2] = 1
    if student_df["time_available"][i].lower().strip() == "20h-22h":
      std_time_slot[student_index][day][3] = 1

  return std_time_slot


## teacher available time.
def get_teacher_time(csv_path):
  teacher_df = pd.read_csv(csv_path)

  teacher_time_slot = []

  for _ in range(NUM_TEACHER):
    time_slot = []
    for __ in range(DAY_OF_WEEK):
      time_slot.append([0, 0, 0, 0])  # 4 shift per day.

    teacher_time_slot.append(time_slot)

  for i in range(len(teacher_df)):
    teacher_index = int(teacher_df["Teacher"][i].split("_")[1]) - 1

    day = int(teacher_df["day_of_week"][i]) - 2  # convert from day of week to index day

    if teacher_df["time_available"][i].lower().strip() == "14h-16h":
      teacher_time_slot[teacher_index][day][0] = 1
    if teacher_df["time_available"][i].lower().strip() == "16h-18h":
      teacher_time_slot[teacher_index][day][1] = 1
    if teacher_df["time_available"][i].lower().strip() == "18h-20h":
      teacher_time_slot[teacher_index][day][2] = 1
    if teacher_df["time_available"][i].lower().strip() == "20h-22h":
      teacher_time_slot[teacher_index][day][3] = 1

  return teacher_time_slot


# Grade
## Student grade
def get_student_grade(student_csv_path):
  student_df = pd.read_csv(student_csv_path)

  std_grade = []
  for _ in range(NUM_STUDENT):
    std_grade.append([0, 0, 0, 0])

  for i in range(len(student_df["Student"])):
    student_index = int(student_df["Student"][i].split("_")[1]) - 1

    if int(student_df["grade"][i]) == 6:
      std_grade[student_index][0] = 1
    if int(student_df["grade"][i]) == 7:
      std_grade[student_index][1] = 1
    if int(student_df["grade"][i]) == 8:
      std_grade[student_index][2] = 1
    if int(student_df["grade"][i]) == 9:
      std_grade[student_index][3] = 1

  return std_grade


## Teacher grade
def get_teacher_grade(teacher_csv_path):
  teacher_df = pd.read_csv(teacher_csv_path)

  teacher_grade = []
  for _ in range(NUM_TEACHER):
    teacher_grade.append([0, 0, 0, 0])

  for i in range(len(teacher_df["Teacher"])):
    student_index = int(teacher_df["Teacher"][i].split("_")[1]) - 1

    if int(teacher_df["grade"][i]) == 6:  # teacher handle +- 1 grade diff
      teacher_grade[student_index][0] = 1
      teacher_grade[student_index][1] = 1

    if int(teacher_df["grade"][i]) == 7:
      teacher_grade[student_index][0] = 1
      teacher_grade[student_index][1] = 1
      teacher_grade[student_index][2] = 1

    if int(teacher_df["grade"][i]) == 8:
      teacher_grade[student_index][1] = 1
      teacher_grade[student_index][2] = 1
      teacher_grade[student_index][3] = 1

    if int(teacher_df["grade"][i]) == 9:
      teacher_grade[student_index][2] = 1
      teacher_grade[student_index][3] = 1
  return teacher_grade


## refer teacher.
def get_ref_teacher(csv_path):
  student_df = pd.read_csv(csv_path)

  ref_teacher = []
  for _ in range(NUM_STUDENT):
    teacher_list = []
    for _ in range(NUM_TEACHER):
      teacher_list.append(0)
    ref_teacher.append(teacher_list)

  for i in range(len(student_df["refer teacher"])):
    teacher_index = int(student_df["refer teacher"][i].split("_")[1]) - 1
    student_index = int(student_df["Student"][i].split("_")[1]) - 1

    ref_teacher[student_index][teacher_index] = 1

  return ref_teacher


def scheduling(result_path, **constraint):
  all_students = range(NUM_STUDENT)
  all_teachers = range(NUM_TEACHER)
  all_subjects = range(NUM_SUBJECT)
  all_grades = range(NUM_GRADE)
  all_dow = range(DAY_OF_WEEK)
  all_shifts = range(NUM_SHIFT)

  model = cp_model.CpModel()
  classes = {}
  is_refer_teacher = {}
  teacher_class = {}

  # 1. Create variable.
  for stu in all_students:
    for t in all_teachers:
      for sub in all_subjects:
        for grade in all_grades:
          for day in all_dow:
            for shift in all_shifts:
              classes[(stu, t, sub, grade, day, shift)] = model.NewBoolVar(
                f"shift_{stu}_{t}_{sub}_{grade}_{day}_{shift}")

              is_refer_teacher[(stu, t, sub, grade, day, shift)] = model.NewBoolVar(
                f"is_ref_teacher_{stu}_{t}_{sub}_{grade}_{day}_{shift}"
              )

  for t in all_teachers:
    for sub in all_subjects:
      for grade in all_grades:
        for day in all_dow:
          for shift in all_shifts:
            teacher_class[(t, grade, sub, day, shift)] = model.NewBoolVar(
              f"grade_teacher_{t}_{grade}_{sub}_{day}_{shift}"
            )

  # 2. Constraint
  # Student have one class with one teacher in specific shifts.
  for t in all_teachers:
    for sub in all_subjects:
      for grade in all_grades:
        for day in all_dow:
          for shift in all_shifts:
            model.AddAtMostOne(classes[(stu, t, sub, grade, day, shift)] for stu in all_students)

  ## teacher have a slot at time.
  for t in all_teachers:
    for day in all_dow:
      for shift in all_shifts:
        list_tmp = []
        for sub in all_subjects:
          for grade in all_grades:
            if (constraint["teacher_subject_req"][t][sub] *
                constraint["teacher_time"][t][day][shift] *
                constraint["teacher_grade"][t][grade]):
              list_tmp.append(teacher_class[(t, grade, sub, day, shift)])

        model.Add(sum(list_tmp) <= 1)

  for t in all_teachers:
    for day in all_dow:
      for shift in all_shifts:
        list_tmp = []
        for sub in all_subjects:
          for grade in all_grades:
            list_tmp.append(teacher_class[(t, grade, sub, day, shift)])

        model.Add(sum(list_tmp) <= 1)

  for t in all_teachers:
    for day in all_dow:
      for shift in all_shifts:
        for sub in all_subjects:
          for grade in all_grades:
            model.Add(
              sum(
                constraint["stu_sub_req"][stu][sub] *
                constraint["teacher_subject_req"][t][sub] *
                constraint["student_time"][stu][day][shift] *
                constraint["teacher_time"][t][day][shift] *
                constraint["student_grade"][stu][grade] *
                constraint["teacher_grade"][t][grade] *
                classes[(stu, t, sub, grade, day, shift)]
                for stu in all_students
              ) >= 1
            ).OnlyEnforceIf(teacher_class[(t, grade, sub, day, shift)])

  ## in specific day, shifts, subject. A student learn a slot at a time.
  for day in all_dow:
    for shift in all_shifts:
      for sub in all_subjects:
        for stu in all_students:
          model.Add(
            sum(
              constraint["stu_sub_req"][stu][sub] *
              constraint["teacher_subject_req"][t][sub] *
              constraint["student_time"][stu][day][shift] *
              constraint["teacher_time"][t][day][shift] *
              constraint["student_grade"][stu][grade] *
              constraint["teacher_grade"][t][grade] *
              classes[(stu, t, sub, grade, day, shift)]
              for grade in all_grades
              for t in all_teachers
            ) <= 1
          )

  ## Each class have maximize 12 students.
  NUMBER_STUDENT_PER_CLASS = 12
  for day in all_dow:
    for shift in all_shifts:
      for sub in all_subjects:
        for grade in all_grades:
          for t in all_teachers:
            num_student = []
            for stu in all_students:
              num_student.append(classes[(stu, t, sub, grade, day, shift)])
            model.Add(sum(num_student) <= NUMBER_STUDENT_PER_CLASS)

  ## each part-time teacher have 10 shift per week.
  NUM_SHIFT_PER_WEEK = 10
  for t in all_teachers:
    num_shifts = []
    for stu in all_students:
      for sub in all_subjects:
        for grade in all_grades:
          for day in all_dow:
            for shift in all_shifts:
              num_shifts.append(classes[(stu, t, sub, grade, day, shift)])
    model.Add(sum(num_shifts) <= NUM_SHIFT_PER_WEEK)

  ## prefer teacher
  for stu in all_students:
    for sub in all_subjects:
      for grade in all_grades:
        for day in all_dow:
          for shift in all_shifts:
            for t in all_teachers:
              tmp_teacher = [classes[(stu, t, sub, grade, day, shift)]]
              model.Add(
                sum(tmp_teacher) *
                constraint["ref_teacher"][stu][t] *
                constraint["stu_sub_req"][stu][sub] *
                constraint["teacher_subject_req"][t][sub] *
                constraint["student_time"][stu][day][shift] *
                constraint["teacher_time"][t][day][shift] *
                constraint["student_grade"][stu][grade] *
                constraint["teacher_grade"][t][grade] >= 1).OnlyEnforceIf(
                is_refer_teacher[(stu, t, sub, grade, day, shift)])

  ## Add Objective function.
  model.Maximize(
    sum(
      teacher_class[(t, grade, sub, day, shift)] +

      is_refer_teacher[(stu, t, sub, grade, day, shift)] +

      constraint["stu_sub_req"][stu][sub] *
      constraint["teacher_subject_req"][t][sub] *
      constraint["student_time"][stu][day][shift] *
      constraint["teacher_time"][t][day][shift] *
      constraint["student_grade"][stu][grade] *
      constraint["teacher_grade"][t][grade] *

      classes[(stu, t, sub, grade, day, shift)]
      for stu in all_students
      for t in all_teachers
      for sub in all_subjects
      for day in all_dow
      for shift in all_shifts
      for grade in all_grades
    )
  )

  ## SOLVER
  solver = cp_model.CpSolver()
  solver.Solve(model)
  status = solver.Solve(model)

  print(f"STATUS: {status}")
  num_slot = 0

  resutl_df = pd.DataFrame(columns=["day", "shift", "subject", "grade", "student", "teacher"])

  if status == cp_model.OPTIMAL:
    print('Solution:')
    for day in all_dow:
      print('Day', day)
      for stu in all_students:
        for t in all_teachers:
          for sub in all_subjects:
            for grade in all_grades:
              for shift in all_shifts:
                if (solver.Value(classes[(stu, t, sub, grade, day, shift)]) == 1) & (
                    solver.Value(teacher_class[(t, grade, sub, day, shift)]) == 1):
                  prefer_teacher = constraint["ref_teacher"][stu][t]
                  print(f" *** {day} - {sub} - {grade} - {shift} - {stu} - {t}")
                  print(
                    f"\t Day: {day + 2} - Subject: {SUBJECT_MAPPING[sub]} - Grade: {grade + 6} \
                   - Shift: {SHIFT_MAPPING[shift]} - student: student-{stu + 1} \
                   - Teacher: teacher-{t + 1} - Prefer teacher: {prefer_teacher} - grade teacher: {solver.Value(teacher_class[(t, grade, sub, day, shift)])}")

                  tmp_df = pd.DataFrame([{
                    "day": f"{day + 2}",
                    "subject": f"{SUBJECT_MAPPING[sub]}",
                    "grade": f"{grade + 6}",
                    "shift": f"{SHIFT_MAPPING[shift]}",
                    "student": f"st_{stu + 1}",
                    "teacher": f"Teacher_{t + 1}"
                  }])

                  num_slot += 1

                  resutl_df = pd.concat([resutl_df, tmp_df], ignore_index=True)
  else:
    print("No optimal solution found !")

  if status == cp_model.OPTIMAL:
    print('There have Solution')
    resutl_df.to_csv(result_path)
  else:
    print("No optimal solution found !")


@click.command()
@click.option('--student_csv_path', prompt="path to student csv data",
              default="../../data/mvp/input/student_df_formated.csv")
@click.option('--teacher_csv_path', prompt="path to teacher csv data",
              default="../../data/mvp/input/teacher_df_formated.csv")
@click.option('--result_path', prompt="path to store result", default="../../data/mvp/output/result.csv")
def run(student_csv_path, teacher_csv_path, result_path):
  # Subject request
  stu_sub_req = get_student_subject_req(student_csv_path)
  teacher_subject_req = get_teacher_subject_req(teacher_csv_path)

  # Time available
  student_time = get_std_time(student_csv_path)
  teacher_time = get_teacher_time(teacher_csv_path)

  # Grade
  student_grade = get_student_grade(student_csv_path)
  teacher_grade = get_teacher_grade(teacher_csv_path)

  # ref teacher
  ref_teacher = get_ref_teacher(student_csv_path)

  scheduling(stu_sub_req=stu_sub_req, teacher_subject_req=teacher_subject_req,
             student_time=student_time, teacher_time=teacher_time,
             student_grade=student_grade, teacher_grade=teacher_grade, ref_teacher=ref_teacher, result_path=result_path)

  print("DONE")


if __name__ == '__main__':
  run()
