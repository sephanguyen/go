# NUM_STUDENT = 55
# NUM_TEACHER = 11
# NUM_SUBJECT = 5
# NUM_GRADE = 12
# NUM_DAY = 46
# NUM_SHIFT = 11

NUM_STUDENT = 26
NUM_TEACHER = 15
NUM_SUBJECT = 5
NUM_GRADE = 12
NUM_DAY = 13
NUM_SHIFT = 6

# The primary time = NUM_SHIFT_FIRST_HALF * NUM_DAY_FIRST_HALF + NUM_SHIFT_SECOND_HALF * NUM_DAY_SECOND_HALF
PERCENT_FIRST_HALF_IN_PRIMARY_TIME = 0.8
PERCENT_FIRST_HALF_IN_SECONDARY_TIME = 0.15
PERCENT_SECOND_HALF_IN_PRIMARY_TIME = 0.5
PERCENT_SECOND_HALF_IN_SECONDARY_TIME = 0.5

NUM_DAY_FIRST_HALF = int(PERCENT_FIRST_HALF_IN_PRIMARY_TIME * (NUM_DAY - 1))
NUM_DAY_SECOND_HALF = (NUM_DAY - 1) - NUM_DAY_FIRST_HALF
NUM_SHIFT_FIRST_HALF = int(NUM_SHIFT * PERCENT_FIRST_HALF_IN_PRIMARY_TIME)
NUM_SHIFT_SECOND_HALF = int(NUM_SHIFT * PERCENT_SECOND_HALF_IN_SECONDARY_TIME)

PERCENT_SLOT_AT_PRI = 0.9

SUBJECT_MAPPING = ["math", "english", "literature", "science", "social_science"]
SUBJECT_RATIO = [3, 3, 9, 9, 9]
SHIFT_MAPPING = ["0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11"]

MAX = 10000000

CENTER_CAPACITY = 12 # with bestco data. booth = 4 -> 4*3 = 12 slot

NUM_LEVEL = 4  # combine grade to level => 1,2,3,4,5: elementary; 6,7,8: middle; 10, 11: high school; 12: 12_high_school
LEVEL_MAPPING = ["elemetary", "middle", "high", "12_high"]

ELEMENTARY = ["4", "5", "6", "7", "8", "9", "10", "11"]
MIDDLE = ["0", "1", "2", "3", "4", "8", "9", "10", "11"]
HIGH = ["2", "3", "4", "5", "6", "7", "8", "9"]
HIGH_12 = ["0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10", "11"]
LEVEL_TIME = [ELEMENTARY, MIDDLE, HIGH, HIGH_12]


LIST_HARD_CONSTRAINT = [
  1,  # 1. 90% slot is allocated at the primary time and 10% at the secondary time
  1,  # 2. slot per day per student ≤ slot apply / number day.
  1,  # 3. Student study one slot at a time.
  1,  # 4. The student does not to be allocated slots greater than the number of slots which they registered.
  1,  # 5. The only subjects which is the same ratio be merged together
  1,  # 6. Teacher only handles one slot at a time
  1,  # 7. for each student, the seasonal slot is allocated first then the regular slot.
  0,  # 8. Staff work up to 8 hours per day, hardcode one shift is 1 hour and 10 minutes for break time
  0,  # 9. Applied capacity in each center
]

# soft constraints
WEIGHT_PREFER_TEACHER = 0
WEIGHT_CONTINUOUS_SLOT = 1
WEIGHT_REMAIN_SLOT = 1
