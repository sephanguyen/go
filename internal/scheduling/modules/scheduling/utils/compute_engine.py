from __future__ import annotations

import re
import sys
import os
from typing import Any
import warnings

from google.api_core.extended_operation import ExtendedOperation
from google.cloud import compute_v1

from internal.scheduling.modules.scheduling.utils.startup_script import \
	StartupScript

"""
Source: https://cloud.google.com/compute/docs/samples/compute-instances-create-from-custom-image?hl=en#code-sample
"""


def get_image_from_family(project: str, family: str) -> compute_v1.Image:
	"""
	Retrieve the newest image that is part of a given family in a project.

	Args:
			project: project ID or project number of the Cloud project you want to get image from.
			family: name of the image family you want to get image from.

	Returns:
			An Image object.
	"""
	image_client = compute_v1.ImagesClient()
	# List of public operating system (OS) images: https://cloud.google.com/compute/docs/images/os-details
	newest_image = image_client.get_from_family(project=project, family=family)
	return newest_image


def disk_from_image(
		disk_type: str,
		disk_size_gb: int,
		boot: bool,
		source_image: str,
		auto_delete: bool = True,
) -> compute_v1.AttachedDisk:
	"""
	Create an AttachedDisk object to be used in VM instance creation. Uses an image as the
	source for the new disk.

	Args:
			 disk_type: the type of disk you want to create. This value uses the following format:
					"zones/{zone}/diskTypes/(pd-standard|pd-ssd|pd-balanced|pd-extreme)".
					For example: "zones/us-west3-b/diskTypes/pd-ssd"
			disk_size_gb: size of the new disk in gigabytes
			boot: boolean flag indicating whether this disk should be used as a boot disk of an instance
			source_image: source image to use when creating this disk. You must have read access to this disk. This can be one
					of the publicly available images or an image from one of your projects.
					This value uses the following format: "projects/{project_name}/global/images/{image_name}"
			auto_delete: boolean flag indicating whether this disk should be deleted with the VM that uses it

	Returns:
			AttachedDisk object configured to be created using the specified image.
	"""
	boot_disk = compute_v1.AttachedDisk()
	initialize_params = compute_v1.AttachedDiskInitializeParams()
	initialize_params.source_image = source_image
	initialize_params.disk_size_gb = disk_size_gb
	initialize_params.disk_type = disk_type
	boot_disk.initialize_params = initialize_params
	# Remember to set auto_delete to True if you want the disk to be deleted when you delete
	# your VM instance.
	boot_disk.auto_delete = auto_delete
	boot_disk.boot = boot
	return boot_disk


def wait_for_extended_operation(
		operation: ExtendedOperation, verbose_name: str = "operation",
		timeout: int = 300
) -> Any:
	"""
	Waits for the extended (long-running) operation to complete.

	If the operation is successful, it will return its result.
	If the operation ends with an error, an exception will be raised.
	If there were any warnings during the execution of the operation
	they will be printed to sys.stderr.

	Args:
			operation: a long-running operation you want to wait on.
			verbose_name: (optional) a more verbose name of the operation,
					used only during error and warning reporting.
			timeout: how long (in seconds) to wait for operation to finish.
					If None, wait indefinitely.

	Returns:
			Whatever the operation.result() returns.

	Raises:
			This method will raise the exception received from `operation.exception()`
			or RuntimeError if there is no exception set, but there is an `error_code`
			set for the `operation`.

			In case of an operation taking longer than `timeout` seconds to complete,
			a `concurrent.futures.TimeoutError` will be raised.
	"""
	result = operation.result(timeout=timeout)
	
	if operation.error_code:
		print(
				f"Error during {verbose_name}: [Code: {operation.error_code}]: {operation.error_message}",
				file=sys.stderr,
				flush=True,
		)
		print(f"Operation ID: {operation.name}", file=sys.stderr, flush=True)
		raise operation.exception() or RuntimeError(operation.error_message)
	
	if operation.warnings:
		print(f"Warnings during {verbose_name}:\n", file=sys.stderr, flush=True)
		for warning in operation.warnings:
			print(f" - {warning.code}: {warning.message}", file=sys.stderr,
			      flush=True)
	
	return result


def create_instance(
		project_id: str,
		zone: str,
		instance_name: str,
		disks: list[compute_v1.AttachedDisk],
		machine_type: str = "n1-standard-1",
		network_link: str = "global/networks/default",
		subnetwork_link: str = None,
		internal_ip: str = None,
		external_access: bool = False,
		external_ipv4: str = None,
		accelerators: list[compute_v1.AcceleratorConfig] = None,
		preemptible: bool = False,
		spot: bool = False,
		instance_termination_action: str = "STOP",
		custom_hostname: str = None,
		delete_protection: bool = False,
		service_account: bool = True,
		startup_script: StartupScript = None
) -> compute_v1.Instance:
	"""
	Send an instance creation request to the Compute Engine API and wait for it to complete.

	Args:
			project_id: project ID or project number of the Cloud project you want to use.
			zone: name of the zone to create the instance in. For example: "us-west3-b"
			instance_name: name of the new virtual machine (VM) instance.
			disks: a list of compute_v1.AttachedDisk objects describing the disks
					you want to attach to your new instance.
			machine_type: machine type of the VM being created. This value uses the
					following format: "zones/{zone}/machineTypes/{type_name}".
					For example: "zones/europe-west3-c/machineTypes/f1-micro"
			network_link: name of the network you want the new instance to use.
					For example: "global/networks/default" represents the network
					named "default", which is created automatically for each project.
			subnetwork_link: name of the subnetwork you want the new instance to use.
					This value uses the following format:
					"regions/{region}/subnetworks/{subnetwork_name}"
			internal_ip: internal IP address you want to assign to the new instance.
					By default, a free address from the pool of available internal IP addresses of
					used subnet will be used.
			external_access: boolean flag indicating if the instance should have an external IPv4
					address assigned.
			external_ipv4: external IPv4 address to be assigned to this instance. If you specify
					an external IP address, it must live in the same region as the zone of the instance.
					This setting requires `external_access` to be set to True to work.
			accelerators: a list of AcceleratorConfig objects describing the accelerators that will
					be attached to the new instance.
			preemptible: boolean value indicating if the new instance should be preemptible
					or not. Preemptible VMs have been deprecated and you should now use Spot VMs.
			spot: boolean value indicating if the new instance should be a Spot VM or not.
			instance_termination_action: What action should be taken once a Spot VM is terminated.
					Possible values: "STOP", "DELETE"
			custom_hostname: Custom hostname of the new VM instance.
					Custom hostnames must conform to RFC 1035 requirements for valid hostnames.
			delete_protection: boolean value indicating if the new virtual machine should be
					protected against deletion or not.
	Returns:
			Instance object.
	"""
	instance_client = compute_v1.InstancesClient()
	
	# Use the network interface provided in the network_link argument.
	network_interface = compute_v1.NetworkInterface()
	network_interface.network = network_link
	if subnetwork_link:
		network_interface.subnetwork = subnetwork_link
	
	if internal_ip:
		network_interface.network_i_p = internal_ip
	
	if external_access:
		access = compute_v1.AccessConfig()
		access.type_ = compute_v1.AccessConfig.Type.ONE_TO_ONE_NAT.name
		access.name = "External NAT"
		access.network_tier = access.NetworkTier.PREMIUM.name
		if external_ipv4:
			access.nat_i_p = external_ipv4
		network_interface.access_configs = [access]
	
	# Collect information into the Instance object.
	instance = compute_v1.Instance()
	instance.network_interfaces = [network_interface]
	instance.name = instance_name
	instance.disks = disks
	if re.match(r"^zones/[a-z\d\-]+/machineTypes/[a-z\d\-]+$", machine_type):
		instance.machine_type = machine_type
	else:
		instance.machine_type = f"zones/{zone}/machineTypes/{machine_type}"
	
	instance.scheduling = compute_v1.Scheduling()
	if accelerators:
		instance.guest_accelerators = accelerators
		instance.scheduling.on_host_maintenance = (
			compute_v1.Scheduling.OnHostMaintenance.TERMINATE.name
		)
	
	if preemptible:
		# Set the preemptible setting
		warnings.warn(
				"Preemptible VMs are being replaced by Spot VMs.", DeprecationWarning
		)
		instance.scheduling = compute_v1.Scheduling()
		instance.scheduling.preemptible = True
	
	if spot:
		# Set the Spot VM setting
		instance.scheduling.provisioning_model = (
			compute_v1.Scheduling.ProvisioningModel.SPOT.name
		)
		instance.scheduling.instance_termination_action = instance_termination_action
	
	if custom_hostname is not None:
		# Set the custom hostname for the instance
		instance.hostname = custom_hostname
	
	if delete_protection:
		# Set the delete protection bit
		instance.deletion_protection = True
	
	# Service account and IAM
	if service_account:
		env = os.environ["ENV"]
		org = os.environ["ORG"]
		service_account_email = f"{env}-auto-scheduling@{project_id}.iam.gserviceaccount.com"
		sa = compute_v1.ServiceAccount(email=service_account_email)
		instance.service_accounts = [sa]
		
		
	# startup script
	metadata = compute_v1.Metadata()
	items = compute_v1.types.Items()
	items.key = "startup-script"
	items.value = startup_script.get_script()
	metadata.items = [items]
	instance.metadata = metadata
	
	# network tags:
	tags = compute_v1.Tags(items=['scheduling'])
	instance.tags = tags
	
	# Prepare the request to insert an instance.
	request = compute_v1.InsertInstanceRequest()
	request.zone = zone
	request.project = project_id
	request.instance_resource = instance
	
	# Wait for the create operation to complete.
	print(f"Creating the {instance_name} instance in {zone}...")
	
	operation = instance_client.insert(request=request)
	
	wait_for_extended_operation(operation, "instance creation")
	
	print(f"Instance {instance_name} created.")
	return instance_client.get(project=project_id, zone=zone,
	                           instance=instance_name)


def create_from_custom_image(
		project_id: str, zone: str, instance_name: str, custom_image_link: str,
		network_link: str,
		startup_script: StartupScript = None
) -> compute_v1.Instance:
	"""
	Create a new VM instance with custom image used as its boot disk.

	Args:
			project_id: project ID or project number of the Cloud project you want to use.
			zone: name of the zone to create the instance in. For example: "us-west3-b"
			instance_name: name of the new virtual machine (VM) instance.
			custom_image_link: link to the custom image you want to use in the form of:
					"projects/{project_name}/global/images/{image_name}"

	Returns:
			Instance object.
	"""
	disk_type = f"zones/{zone}/diskTypes/pd-standard"
	disks = [disk_from_image(disk_type, 50, True, custom_image_link, True)]
	instance = create_instance(project_id=project_id, zone=zone,
	                           instance_name=instance_name, disks=disks,
	                           network_link=network_link,
	                           external_access=True,
	                           machine_type="n2-standard-32",
	                           startup_script=startup_script)
	return instance


def create_instance_from_template(
		project_id: str, zone: str, instance_name: str, instance_template_url: str
) -> compute_v1.Instance:
	"""
	Creates a Compute Engine VM instance from an instance template.

	Args:
			project_id: ID or number of the project you want to use.
			zone: Name of the zone you want to check, for example: us-west3-b
			instance_name: Name of the new instance.
			instance_template_url: URL of the instance template used for creating the new instance.
					It can be a full or partial URL.
					Examples:
					- https://www.googleapis.com/compute/v1/projects/project/global/instanceTemplates/example-instance-template
					- projects/project/global/instanceTemplates/example-instance-template
					- global/instanceTemplates/example-instance-template

	Returns:
			Instance object.
	"""
	instance = compute_v1.Instance()
	
	instance.name = instance_name
	
	# startup script
	metadata = compute_v1.Metadata()
	items = compute_v1.types.Items()
	items.key = "startup-script"
	items.value = """
        #!/bin/bash

        sudo echo "Hello world"
        """
	
	metadata.items = [items]
	instance.metadata = metadata
	
	instance_client = compute_v1.InstancesClient()
	
	instance_insert_request = compute_v1.InsertInstanceRequest()
	instance_insert_request.project = project_id
	instance_insert_request.zone = zone
	instance_insert_request.source_instance_template = instance_template_url
	instance_insert_request.instance_resource = instance
	
	operation = instance_client.insert(instance_insert_request)
	wait_for_extended_operation(operation, "instance creation")
	
	return instance_client.get(project=project_id, zone=zone,
	                           instance=instance_name)


def create_instance_from_template_with_overrides(
		project_id: str,
		zone: str,
		instance_name: str,
		instance_template_name: str,
		machine_type: str,
		new_disk_source_image: str,
) -> compute_v1.Instance:
	"""
	Creates a Compute Engine VM instance from an instance template, changing the machine type and
	adding a new disk created from a source image.

	Args:
			project_id: ID or number of the project you want to use.
			zone: Name of the zone you want to check, for example: us-west3-b
			instance_name: Name of the new instance.
			instance_template_name: Name of the instance template used for creating the new instance.
			machine_type: Machine type you want to set in following format:
					"zones/{zone}/machineTypes/{type_name}". For example:
					- "zones/europe-west3-c/machineTypes/f1-micro"
					- You can find the list of available machine types using:
						https://cloud.google.com/sdk/gcloud/reference/compute/machine-types/list
			new_disk_source_image: Path the the disk image you want to use for your new
					disk. This can be one of the public images
					(like "projects/debian-cloud/global/images/family/debian-10")
					or a private image you have access to.
					For a list of available public images, see the documentation:
					http://cloud.google.com/compute/docs/images

	Returns:
			Instance object.
	"""
	instance_client = compute_v1.InstancesClient()
	instance_template_client = compute_v1.InstanceTemplatesClient()
	
	# Retrieve an instance template by name.
	instance_template = instance_template_client.get(
			project=project_id, instance_template=instance_template_name
	)
	
	# Adjust diskType field of the instance template to use the URL formatting required by instances.insert.diskType
	# For instance template, there is only a name, not URL.
	for disk in instance_template.properties.disks:
		if disk.initialize_params.disk_type:
			disk.initialize_params.disk_type = (
				f"zones/{zone}/diskTypes/{disk.initialize_params.disk_type}"
			)
	
	instance = compute_v1.Instance()
	instance.name = instance_name
	instance.machine_type = machine_type
	instance.disks = list(instance_template.properties.disks)
	
	new_disk = compute_v1.AttachedDisk()
	new_disk.initialize_params.disk_size_gb = 50
	new_disk.initialize_params.source_image = new_disk_source_image
	new_disk.auto_delete = True
	new_disk.boot = False
	new_disk.type_ = "PERSISTENT"
	
	instance.disks.append(new_disk)
	
	instance_insert_request = compute_v1.InsertInstanceRequest()
	instance_insert_request.project = project_id
	instance_insert_request.zone = zone
	instance_insert_request.instance_resource = instance
	instance_insert_request.source_instance_template = instance_template.self_link
	
	operation = instance_client.insert(instance_insert_request)
	wait_for_extended_operation(operation, "instance creation")
	
	return instance_client.get(project=project_id, zone=zone,
	                           instance=instance_name)

