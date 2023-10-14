# Readme


## Instruction
Aphelios is grpc service include 3 module for omr detection: question_detector, id_question_detector and answer_detector. 
The more detail in [OMR-proof of concept](https://manabie.atlassian.net/wiki/spaces/TECH/pages/452526629/Phase+1+Proof+of+concept)


1. build docker file:
   from ./backend. 
    ```
   docker build -t asia.gcr.io/student-coach-e1e95/aphelios:20220616 -f ./backend/developments/python.Dockerfile .
   ```
