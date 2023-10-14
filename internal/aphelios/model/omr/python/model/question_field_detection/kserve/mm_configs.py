# the new config inherits the base configs to highlight the necessary modification
_base_ = "/mmrotate/configs/rotated_retinanet/rotated_retinanet_obb_r50_fpn_1x_dota_oc.py" # <===== edit this base path. 

# 1. dataset settings
dataset_type = 'DOTADataset'
classes = ('question_field', )
angle_version = 'oc'
img_norm_cfg = dict(
    mean=[123.675, 116.28, 103.53], std=[58.395, 57.12, 57.375], to_rgb=True)

train_pipeline = [
    dict(type='LoadImageFromFile'),
    dict(type='LoadAnnotations', with_bbox=True),
    dict(type='RResize', img_scale=(1024, 1024)),
    dict(
        type='RRandomFlip',
        flip_ratio=[0.25, 0.25, 0.25],
        direction=['horizontal', 'vertical', 'diagonal'],
        version=angle_version),
    dict(type='Normalize', **img_norm_cfg),
    dict(type='Pad', size_divisor=32),
    dict(type='DefaultFormatBundle'),
    dict(type='Collect', keys=['img', 'gt_bboxes', 'gt_labels'])
] 

dataset_train = dict(
        type='RepeatDataset',
        times=500,
        dataset=dict(  # This is the original config of Dataset_A
            type=dataset_type,
            classes=classes,
            ann_file='/mnt/disks/attach-disk/OMR/vongho/omr-/model/cornerdetection/dota_format_question_field/train',
            img_prefix='/mnt/disks/attach-disk/OMR/vongho/omr-/data_for_trainning/id_question_field/id_question_field/images/train',
            pipeline=train_pipeline
        )
    )

data = dict(
  samples_per_gpu=2,
  workers_per_gpu=2,
  train=dataset_train,
  val=dict(
    type=dataset_type,
    # explicitly add your class names to the field `classes`
    classes=classes,
    ann_file='/mnt/disks/attach-disk/OMR/vongho/omr-/model/cornerdetection/dota_format_question_field/dev',
    img_prefix='/mnt/disks/attach-disk/OMR/vongho/omr-/data_for_trainning/id_question_field/id_question_field/images/dev'),
  test=dict(
    type=dataset_type,
    # explicitly add your class names to the field `classes`
    classes=classes,
    ann_file='/mnt/disks/attach-disk/OMR/vongho/omr-/model/cornerdetection/dota_format_question_field/dev',
    img_prefix='/mnt/disks/attach-disk/OMR/vongho/omr-/data_for_trainning/id_question_field/id_question_field/images/dev'))

# 2. model settings

model = dict(
    type='RotatedRetinaNet',
    backbone=dict(
        type='ResNet',
        depth=50,
        num_stages=4,
        out_indices=(0, 1, 2, 3),
        frozen_stages=1,
        zero_init_residual=False,
        norm_cfg=dict(type='BN', requires_grad=True),
        norm_eval=True,
        style='pytorch',
        init_cfg=dict(type='Pretrained', checkpoint='torchvision://resnet50')),
    neck=dict(
        type='FPN',
        in_channels=[256, 512, 1024, 2048],
        out_channels=256,
        start_level=1,
        add_extra_convs='on_input',
        num_outs=5),
    bbox_head=dict(
        type='RotatedRetinaHead',
        num_classes=1,
        in_channels=256,
        stacked_convs=4,
        feat_channels=256,
        assign_by_circumhbbox=None,
        anchor_generator=dict(
            type='RotatedAnchorGenerator',
            octave_base_scale=4,
            scales_per_octave=3,
            ratios=[1.0, 0.5, 2.0],
            strides=[8, 16, 32, 64, 128]),
        bbox_coder=dict(
            type='DeltaXYWHAOBBoxCoder',
            angle_range=angle_version,
            norm_factor=None,
            edge_swap=False,
            proj_xy=False,
            target_means=(.0, .0, .0, .0, .0),
            target_stds=(1.0, 1.0, 1.0, 1.0, 1.0)),
        loss_cls=dict(
            type='FocalLoss',
            use_sigmoid=True,
            gamma=2.0,
            alpha=0.25,
            loss_weight=1.0),
        loss_bbox=dict(type='L1Loss', loss_weight=1.0)),
    train_cfg=dict(
        assigner=dict(
            type='MaxIoUAssigner',
            pos_iou_thr=0.5,
            neg_iou_thr=0.4,
            min_pos_iou=0,
            ignore_iof_thr=-1,
            iou_calculator=dict(type='RBboxOverlaps2D')),
        allowed_border=-1,
        pos_weight=-1,
        debug=False),
    test_cfg=dict(
        nms_pre=2000,
        min_bbox_size=0,
        score_thr=0.05,
        nms=dict(iou_thr=0.8),
        max_per_img=2000))

log_config = dict(
    exp_name="question-field-detection",
    interval=50,
    hooks=[
        dict(type='MlflowLoggerHook')
    ])

work_dir = '/mnt/disks/attach-disk/OMR/vongho/omr-/model/cornerdetection/workdir_question_field/rotated_retinanet_hbb_r50_fpn_1x_dota_oc'  # Directory to save the model checkpoints and logs for the current experiments.

