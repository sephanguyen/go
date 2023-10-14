#!/bin/python3

import os
import yaml

CURRENT_DIR = os.path.dirname(os.path.abspath(__file__))
METADATA_DIR = os.path.join(
    CURRENT_DIR, "../deployments/helm/manabie-all-in-one/charts/bob/files/hasura/metadata")
TABLE_FILE = os.path.join(METADATA_DIR, "tables.yaml")
OUTPUT_TABLE_FILE = os.path.join(METADATA_DIR, "tables_stg.yaml")
FUNCTION_FILE = os.path.join(METADATA_DIR, "functions.yaml")
OUTPUT_FUNCTION_FILE = os.path.join(METADATA_DIR, "functions_stg.yaml")

TBL_BLACKLIST = set([
    "course_teaching_time",
    "course_location_schedule",
    "classroom",
    "lesson_groups",
    "lesson_members",
    "lessons",
    "lessons_teachers",
    "lessons_courses",
    "lesson_classrooms",
    "lesson_room_states",
    "reallocation",
    "lesson_recorded_videos",
    "zoom_account",
    "lesson_report_details",
    "lesson_reports",
    "lesson_student_subscription_access_path",
    "lesson_student_subscriptions",
    "lesson_polls",
    "lesson_members_states",
    "live_lesson_conversation",
    "live_lesson_sent_notifications",
    "live_room",
    "live_room_activity_logs",
    "live_room_log",
    "live_room_member_state",
    "live_room_poll",
    "live_room_recorded_videos",
    "live_room_state",
    "virtual_classroom_log",
    "partner_dynamic_form_field_values",
    "partner_form_configs",
    "lesson_schedules",
])

FN_BLACKLIST = [
    "get_all_group_report_by_lesson_id",
    "get_all_individual_report_of_student",
    "get_all_individual_report_of_student_v2",
    "get_partner_dynamic_form_field_values_by_student",
    "get_previous_lesson_report_group",
    "get_previous_report_of_student",
    "get_previous_report_of_student_v2",
    "get_previous_report_of_student_v3",
    "get_previous_report_of_student_v4",
]


def process_relationship(rls):
    if len(rls) == 0:
        raise Exception('input relationship cannot be empty')

    p_rls = []
    for ar in rls:
        # Check foreign keys
        try:
            fk_tblname = ar['using']['foreign_key_constraint_on']['table']['name']
            if fk_tblname in TBL_BLACKLIST:
                continue
        except Exception:
            pass

        # Check manual config
        try:
            mc_tblname = ar['using']['manual_configuration']['remote_table']['name']
            if mc_tblname in TBL_BLACKLIST:
                continue
        except Exception:
            pass

        p_rls.append(ar)

    return p_rls


def process_tables_yaml(input_path, output_path=None):
    if output_path is None:
        output_path = input_path

    with open(input_path, 'r') as f:
        data = yaml.safe_load(f)

    p_data: list = []
    for v in data:
        tblname = v.get('table', {}).get('name', '')
        # discard tables with matching names
        if tblname in TBL_BLACKLIST:
            continue

        if 'array_relationships' in v:
            v['array_relationships'] = process_relationship(
                v['array_relationships'])
        if 'object_relationships' in v:
            v['object_relationships'] = process_relationship(
                v['object_relationships'])
        p_data.append(v)

    with open(output_path, 'w') as f:
        yaml.safe_dump(p_data, f, sort_keys=False)


def process_functions_yaml(input_path, output_path=None):
    if output_path is None:
        output_path = input_path

    with open(input_path, 'r') as f:
        data = yaml.safe_load(f)

    p_data = []
    for v in data:
        fnname = v.get('function', {}).get('name', '')
        if fnname in FN_BLACKLIST:
            continue
        p_data.append(v)

    with open(output_path, 'w') as f:
        yaml.safe_dump(p_data, f, sort_keys=False)


if __name__ == '__main__':
    process_tables_yaml(TABLE_FILE, OUTPUT_TABLE_FILE)
    process_functions_yaml(FUNCTION_FILE, OUTPUT_FUNCTION_FILE)
