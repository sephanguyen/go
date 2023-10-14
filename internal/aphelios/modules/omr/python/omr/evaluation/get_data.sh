#!bin/bash

# 1. login google
#gcloud auth login

# 2. set projects dev-manabie-online
gcloud config set project dev-manabie-online

# 2. Pull testing data.
mkdir -p testing_data
gsutil -m rsync -r gs://dev-manabie-data/machine_learning/omr/test_set/answer_sheet ./testing_data/