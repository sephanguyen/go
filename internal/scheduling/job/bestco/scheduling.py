import sys

import click

sys.path += ["../../constraints", "./internal/scheduling/constraints", "./constraints"]
from constraint import Constraint

import time
import pandas as pd

from ortools.sat.python import cp_model
from config import NUM_TEACHER, NUM_DAY, NUM_STUDENT, NUM_SHIFT, SHIFT_MAPPING,\
  SUBJECT_MAPPING, WEIGHT_CONTINUOUS_SLOT, NUM_LEVEL, LEVEL_MAPPING, \
  WEIGHT_PREFER_TEACHER, WEIGHT_REMAIN_SLOT, LIST_HARD_CONSTRAINT
from utils import get_stu_level, get_stu_sub, get_stu_avail_day, \
  get_stu_pre_teacher, get_teacher_sub_level, get_all_stu_slot, \
  get_stu_slot_per_day, get_all_stu_sub_slot, get_merge_subject, \
  add_merged_subject_for_teacher, add_merged_subject_for_student, \
  post_process_assign_slot_types, post_process_check_prefer_teacher, \
  get_prefer_score, allocate_subject_slot, get_stu_avail_time, \
  get_teacher_available_time


def scheduling(result_path, **constraint):
  all_students = range(NUM_STUDENT)
  all_teachers = range(NUM_TEACHER)
  all_subjects = range(len(constraint["merge_sub_list"]))
  all_level = range(NUM_LEVEL)
  all_day = range(NUM_DAY)
  all_shifts = range(NUM_SHIFT)

  model = cp_model.CpModel()
  classes = {}
  is_refer_teacher = {}
  primary_slot = {}
  teacher_slot = {}
  teacher_shift = {}
  student_slot = {}
  student_subject_class = {}


  # 1. Create variable.
  var_time = time.time()
  for stu in all_students:
    for t in all_teachers:
      for sub in all_subjects:
        for level in all_level:
          for day in all_day:
            for shift in all_shifts:
              classes[(stu, t, sub, level, day, shift)] = model.NewBoolVar(
                f"shift_{stu}_{t}_{sub}_{level}_{day}_{shift}")

              is_refer_teacher[(stu, t, sub, level, day, shift)] = model.NewBoolVar(
                f"is_ref_teacher_{stu}_{t}_{sub}_{level}_{day}_{shift}"
              )

  for t in all_teachers:
    for sub in all_subjects:
      for day in all_day:
        for shift in all_shifts:
          for level in all_level:
            teacher_slot[(t, sub, day, shift, level)] = model.NewBoolVar(
              f"teacher_slot_{t}_{sub}_{day}_{shift}_{level}"
            )

  for t in all_teachers:
    for day in all_day:
      for shift in all_shifts:
        teacher_shift[(t, day, shift)] = model.NewBoolVar(
          f"teacher_shift_{t}_{day}_{shift}"
        )

  for stu in all_students:
    for day in all_day:
      for shift in all_shifts:
        student_slot[(stu, day, shift)] = model.NewBoolVar(
          f"student_slot_{stu}_{day}_{shift}"
        )

  for stu in all_students:
    for day in all_day:
      for shift in all_shifts:
        for sub in all_subjects:
          student_subject_class[(stu, day, shift, sub)] = model.NewBoolVar(
            f"student_sub_class_{stu}_{day}_{shift}_{sub}"
          )

  print(f"create variable take:{time.time()-var_time}")

  #############################
  ##### constraints list ######
  #############################
  cnstr_time = time.time()
  print("Adding constraints...")

  constraintList = Constraint(constraint=constraint, model=model, list_of_constraints=LIST_HARD_CONSTRAINT)

  constraintList.add_soft_cnstr_consecutive_subject_teacher(classes=classes)

  model = constraintList.get_model()

  print(f"Adding constraints finish, take: {time.time()-cnstr_time}")

  #################################################
  ########### Solver ##############################
  #################################################

  ## Add Objective function.
  ## SOLVER
  solver = cp_model.CpSolver()
  solver.parameters.num_search_workers = 8

  #######################################################
  # Objective values 1: maximize the avail class with the ratio in each class
  sol_time = time.time()
  print("Solve the first objective function ...")

  total_avail_class = sum(
    # continuos slot
    0 if ((shift + 1) >= NUM_SHIFT)
      else WEIGHT_CONTINUOUS_SLOT * student_slot[(stu, day, shift + 1)] +
    0 if ((shift - 1) < 0)
      else WEIGHT_CONTINUOUS_SLOT * student_slot[(stu, day, shift - 1)] +

    # soft constraints prefer teacher
    WEIGHT_PREFER_TEACHER *
    constraint["stu_teacher_score"][stu][t] +

    # soft constraints teacher only teach only subject at a time
    teacher_slot[(t, sub, day, shift, level)] +

    constraint["stu_level"][stu][level] *
    constraint["stu_sub"][stu][sub] *
    constraint["teacher_sub_level"][t][sub][
        level] *
    constraint["stu_time"][stu][day][shift] *
    constraint["teacher_time"][t][day][shift] *
    classes[(stu, t, sub, level, day, shift)]
      for stu in all_students
      for t in all_teachers
      for sub in all_subjects
      for day in all_day
      for shift in all_shifts
      for level in all_level
  )

  ## the remain slot
  total_remain_slot_objective_values = sum(
    # slot applied of student for that sub
    -
    sum(
      classes[(stu, t, sub, level, day, shift)] if
      SUBJECT_MAPPING[sub] in constraint["merge_sub_list"][sub_merge].split("+") else 0
      for t in all_teachers
      for level in all_level
      for day in all_day
      for shift in all_shifts
      for sub_merge in all_subjects
    )
    +
    # sum slot
    constraint["slot_of_each_subject"][stu][sub]
    for stu in all_students
    for sub in range(len(SUBJECT_MAPPING))
  )

  ## number student in each slot
  total_num_student_in_each_slot = sum(
    [
      sum([classes[(stu2, t, sub, level, day, shift)]
                for stu2 in all_students]
      )
      for t in all_teachers
      for sub in all_subjects
      for level in all_level
      for day in all_day
      for shift in all_shifts
  ])

  model.Maximize(WEIGHT_REMAIN_SLOT * total_remain_slot_objective_values
                  + total_num_student_in_each_slot)  # maximize the number student in each class.

  status = solver.Solve(model)
  print(f"Finish Solve the first objective function, take {time.time()-sol_time}")

  #######################################################
  # Objective values 2: maximize the avail class with the ratio in each class < ratio subject

  ### Hint (speed up solving)
  hint_time = time.time()
  print(f"Add hint result.")

  for stu in all_students:
    for t in all_teachers:
      for sub in all_subjects:
        for day in all_day:
          for shift in all_shifts:
            for level in all_level:
              model.AddHint(classes[(stu, t, sub, level, day, shift)],
                            solver.Value(classes[(stu, t, sub, level, day, shift)]))

  for day in all_day:
    for shift in all_shifts:
      for sub_relate in all_subjects:
        for stu in all_students:
          model.AddHint(student_subject_class[(stu, day, shift, sub_relate)],
                        solver.Value(student_subject_class[(stu, day, shift, sub_relate)]))

  print(f"Finish, add hint time take {time.time()-hint_time}")

  ## Run the second objective function base on the result from the first solution
  ### each slot, number student in each slot < ratio(slot 's subject)
  sol_time_2 = time.time()
  print(f"Add the second objective function...")

  constraintList2 = Constraint(constraint=constraint, model=model, list_of_constraints=LIST_HARD_CONSTRAINT)


  constraintList2.add_cnstr_num_stu_with_ratio(classes=classes)
  constraintList2.add_cnstr_90_primary_time(primary_slot=primary_slot, classes=classes)
  constraintList2.add_cnstr_max_slot_per_day(classes=classes)
  constraintList2.add_cnstr_merge_subject_with_the_same_ratio(classes=classes, teacher_slot=teacher_slot,
                                                             student_subject_class=student_subject_class)
  constraintList2.add_cnstr_student_study_one_slot_at_a_time(classes=classes)
  constraintList2.add_cnstr_teacher_only_teach_one_slot_at_a_time(teacher_slot=teacher_slot)
  constraintList2.add_soft_cnstr_consecutive_slot(student_slot=student_slot, classes=classes)
  constraintList2.add_cnstr_max_shift_per_day_staff(teacher_shift=teacher_shift, classes=classes)
  constraintList2.add_cnstr_center_capacity(classes=classes)


  model.Maximize(total_avail_class)  # maximize avail class
  status = solver.Solve(model)
  print(f"Finish solve second objective function. take {time.time()-sol_time_2}")
  print(f"STATUS: {status}")
  num_slot = 0

  resutl_df = pd.DataFrame(columns=["day", "shift", "subject", "level", "student", "teacher"])

  write_time = time.time()
  print(f"Write result to csv...")
  if status == cp_model.OPTIMAL:
    print("there are some solution, the best one will be write on csv file ... ")
    for day in all_day:
      for stu in all_students:
        for t in all_teachers:
          for sub in all_subjects:
            for level in all_level:
              for shift in all_shifts:
                if (solver.Value(classes[(stu, t, sub, level, day, shift)]) == 1):
                  SUBJECT = constraint["merge_sub_list"][sub]
                  is_prime_slot = 0
                  if (shift, day) in primary_slot.keys():
                    is_prime_slot = 1

                  tmp_df = pd.DataFrame([{
                    "day": f"{day + 2}",
                    "subject": f"{SUBJECT}",
                    "level": f"{LEVEL_MAPPING[level]}",
                    "shift": f"{SHIFT_MAPPING[shift]}",
                    "student": f"st_{stu + 1}",
                    "teacher": f"Teacher_{t + 1}",
                    "is_primary_slot": f"{is_prime_slot}"
                  }])
                  num_slot += 1
                  resutl_df = pd.concat([resutl_df, tmp_df], ignore_index=True)
    print("Finish!!! ")
  else:
    print("No optimal solution found !")

  print(f"Finish writing, take:{time.time()-write_time}")

  # Statistics.
  print('\nStatistics')
  print('  - num slot : %i' % num_slot)
  print('  - conflicts: %i' % solver.NumConflicts())
  print('  - branches : %i' % solver.NumBranches())
  print('  - wall time: %f s' % solver.WallTime())
  print('  - objective value: %f s' % solver.ObjectiveValue())

  # post process
  ## get actual subject from merged subject
  resutl_df = allocate_subject_slot(result_df=resutl_df,
                                    slot_of_each_subject=constraint["slot_of_each_subject"])
  resutl_df.to_csv(result_path)

  total_remain_slot_objective_values = 0
  avail_class_objective_values = 0
  return avail_class_objective_values, total_remain_slot_objective_values


@click.command()
@click.option('--teacher_csv_path', prompt="teacher course path:",
              default="./internal/scheduling/data/scheduling/input/teacher_formated.csv", required=True)
@click.option('--student_csv_path', prompt="student course path:",
              default="./internal/scheduling/data/scheduling/input/student_course_formated.csv", required=True)
@click.option('--result_path', prompt="result path:",
              default="./internal/scheduling/data/scheduling/output/result.csv",
              required=True)
def run_scheduling(teacher_csv_path, student_csv_path, result_path):
  start = time.time()

  # grade
  stu_level = get_stu_level(stu_path=student_csv_path)

  # day
  stu_day = get_stu_avail_day(stu_path=student_csv_path)

  # get student time
  list_stu_time = get_stu_avail_time(student_path=student_csv_path)

  # get teacher time
  list_teacher_time = get_teacher_available_time(teacher_path=teacher_csv_path)

  # prefer teacher
  stu_teacher = get_stu_pre_teacher(stu_path=student_csv_path)

  # get max slot of each student (seasonal slot + sd slot)
  stu_slot = get_all_stu_slot(stu_path=student_csv_path)

  # slot per day.
  stu_slot_per_day = get_stu_slot_per_day(stu_day=stu_day, all_stu_slot=stu_slot)

  # slot of student in each subject
  slot_of_each_subject = get_all_stu_sub_slot(stu_path=student_csv_path)

  ##################
  # merge subject
  merge_sub_list, merge_ratio_list, teacher_sub = get_merge_subject(teacher_csv_path)

  # convert merge subject for teacher => only add merged subject if they can teach all of them
  teacher_df = add_merged_subject_for_teacher(teacher_csv_path, teacher_sub)

  # convert merge subject for student => append merge subject also so constraints of merged subject and subjects < sum of subjects
  stu_sub_df = add_merged_subject_for_student(student_csv=student_csv_path, merge_subject_list=merge_sub_list)
  # sub
  stu_sub = get_stu_sub(stu_df=stu_sub_df, subject_list=merge_sub_list)

  teacher_sub_level = get_teacher_sub_level(teacher_df=teacher_df, subject_list=merge_sub_list)

  # prefer teacher of each student
  stu_teacher_score = get_prefer_score(std_path=student_csv_path)

  print(f"Convert take : {time.time() - start}")
  print(f"Create param...")
  start2 = time.time()

  scheduling(stu_level=stu_level, teacher_sub_level=teacher_sub_level,
             stu_sub=stu_sub,
             stu_time=list_stu_time, teacher_time=list_teacher_time,
             stu_teacher=stu_teacher, stu_slot=stu_slot, stu_slot_per_day=stu_slot_per_day,
             slot_of_each_subject=slot_of_each_subject, merge_sub_list=merge_sub_list,
             merge_ratio_list=merge_ratio_list, stu_teacher_score=stu_teacher_score,
             result_path=result_path)

  if LIST_HARD_CONSTRAINT[6] == 1:
    post_process_assign_slot_types(student_csv_path, result_path)
  post_process_check_prefer_teacher(result_path, stu_teacher_score)

  print(f"Scheduling take {time.time() - start2}")
  print("DONE")


if __name__ == '__main__':
  run_scheduling()
