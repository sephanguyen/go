# Instruction 

This repo used from [mmrotate framework](https://github.com/open-mmlab/mmrotate). Then apply the config model from `internal/aphelios/model/omr/python/model/question_field_detection/kserve/mlserve.py`

In this model, also include tool in `./coco2dota.py` for convert data from coco format to dota format which is used for training in *mmrotate*

# How to use

1. clone repo of mmrotate. 
   ```
   git clone git@github.com:open-mmlab/mmrotate.git
   ```

2. For training model
   
    - Change the specific path of _base_ model path in `internal/aphelios/model/omr/python/model/question_field_detection/kserve/mm_configs.py`. _base_ in folder which you have clone. ( `mmrotate/configs/rotated_retinanet/rotated_retinanet_obb_r50_fpn_1x_dota_oc.py` ) 
     

    
    - Run
        ```
        #from mmrotate folder 

        $ python tools/train.py  internal/aphelios/model/omr/python/model/question_field_detection/kserve/mm_configs.py
        ```

3. Visualize the result:
    ```
    python ./tools/test.py \
    ./mm_configs.py \
    checkpoints/SOME_CHECKPOINT.pth \
    --show-dir work_dirs/vis
    ```
   
4. Evaluate the result:
   - Use the same mm_config.py
   - Data test: store on pachyderm.
     - repo: question_field
     - branch: test

# Reference:

[1]. Official mmrotate [doc](https://mmrotate.readthedocs.io/en/latest/install.html#installation)