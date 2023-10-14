import logging
import os

from kubernetes import client
from kubernetes import config

if "KUBERNETES_SERVICE_HOST" in os.environ:
  config.load_incluster_config()
else:
  config.load_kube_config()

default_shared_volume_mount = client.V1VolumeMount(
  name="shared-data",
  mount_path="/service"
)

default_shared_volume = client.V1Volume(
  name=default_shared_volume_mount.name,
  empty_dir={},
)


class Kubernetes:
  def __init__(self):

    # Init Kubernetes
    self.core_api = client.CoreV1Api()
    self.batch_api = client.BatchV1Api()

  def create_namespace(self, namespace):

    namespaces = self.core_api.list_namespace()
    all_namespaces = []
    for ns in namespaces.items:
      all_namespaces.append(ns.metadata.name)

    if namespace in all_namespaces:
      logging.info(f"Namespace {namespace} already exists. Reusing.")
    else:
      namespace_metadata = client.V1ObjectMeta(name=namespace)
      self.core_api.create_namespace(
        client.V1Namespace(metadata=namespace_metadata)
      )
      logging.info(f"Created namespace {namespace}.")

    return namespace


  @staticmethod
  def get_default_image():
    image = "asia-docker.pkg.dev/student-coach-e1e95/manaverse/auto-scheduling-job:2023062100"
    if os.environ["ENV"] == "local":
      image = "localhost:5001/" + image
    return image

  @staticmethod
  def create_convert_csv_container(image, name="convert-csv", pull_policy="IfNotPresent", *volume_mounts: client.V1VolumeMount):
    volume_mounts += (default_shared_volume_mount,)

    container = client.V1Container(
      image=image,
      name=name,
      image_pull_policy=pull_policy,
      command=["sh", "-c", """
      export PYTHONPATH=. && \
      python3 ./script/convert_bestco_format/convert_format.py \
        --student_csv_path=./data/bestco/input/raw/student_available_slot_master.csv \
        --teacher_csv_path=./data/bestco/input/raw/teacher_available_slot_master.csv \
        --teacher_subject_csv=./data/bestco/input/raw/teacher_subject.csv \
        --day_csv_path=./data/bestco/input/raw/center_opening_slot.csv \
        --applied_slot=./data/bestco/input/raw/applied_slot.csv
      """],
      volume_mounts=list(volume_mounts),
    )

    logging.info(
      f"Created container with name: {container.name}, "
      f"image: {container.image} and args: {container.args}"
    )

    return container

  @staticmethod
  def create_scheduling_container(image, name="scheduling", pull_policy="IfNotPresent", *volume_mounts: client.V1VolumeMount):
    volume_mounts += (default_shared_volume_mount,)

    container = client.V1Container(
      image=image,
      name=name,
      image_pull_policy=pull_policy,
      command=["python3", "./job/bestco/scheduling.py",
               "--teacher_csv_path=./data/scheduling/input/teacher_formated.csv",
               "--student_csv_path=./data/scheduling/input/student_course_formated.csv",
               "--result_path=./data/scheduling/output/result_run_job.csv"],
      volume_mounts=list(volume_mounts),
    )

    logging.info(
      f"Created container with name: {container.name}, "
      f"image: {container.image} and args: {container.args}"
    )

    return container

  @staticmethod
  def create_pod_template(pod_name, init_containers: list[client.V1Container], containers: list[client.V1Container], *volumes: client.V1Volume):
    volumes += (default_shared_volume,)

    # default on default config
    spec = client.V1PodSpec(
      restart_policy="Never",
      init_containers=init_containers,
      containers=containers,
      automount_service_account_token=False,
      volumes=list(volumes),
    )

    if os.environ["ENV"] == "stag":
      # specific on staging env first
      spec.node_selector = {"cloud.google.com/gke-nodepool": "e2-standard-8-ml-spot"}
      spec.tolerations = [
        client.V1Toleration(
          key="e2-standard-8-ml-scheduling-timetable-spot",
          operator="Exists",
          effect="NoSchedule"
        )
      ]

    pod_template = client.V1PodTemplateSpec(
      spec=spec,
      metadata=client.V1ObjectMeta(name=pod_name, labels={"pod_name": pod_name, "sidecar.istio.io/inject": "false"}),
    )

    return pod_template

  @staticmethod
  def create_job(job_name, pod_template):
    metadata = client.V1ObjectMeta(name=job_name, labels={"job_name": job_name})

    job = client.V1Job(
      api_version="batch/v1",
      kind="Job",
      metadata=metadata,
      spec=client.V1JobSpec(backoff_limit=0, template=pod_template),
    )

    return job
