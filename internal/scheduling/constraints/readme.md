# Instruction

This folder store code test constraint. Input and result of scheduling job be store at `internal/scheduling/data/scheduling/input` and `internal/scheduling/data/scheduling/output`

# Command:

- At the root folder /backend
  ```
  cd /backend
  ```
- install environment 
  ```
  pip install -r ./internal/scheduling/requirement.txt
  ```
- Run test
  ```
  python ./internal/scheduling/constraints/test_constraint.py
  ```

# List of constraint implemeted

- slot per day per student ≤ slot apply / number day
- 90% slot allocate at the primary time (85% of the first half + 50% of the second half) and 10% allocate at the secondary time (15% of first half + 50% of second half)

# References:

- [List of constraints](https://docs.google.com/document/d/1A5V_MI8LpN5OeojylXKIIwomtQyfF1DDuWs7iEnEArs/edit?usp=sharing)