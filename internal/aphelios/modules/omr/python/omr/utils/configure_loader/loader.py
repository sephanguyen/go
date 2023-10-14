import yaml

from yaml.loader import SafeLoader

class MinioConnector():
  secrect_config_path = ""
  config_path = ""

  def set_secrect_config_path(self, path):
    self.secrect_config_path = path
    return self.secrect_config_path

  def set_configure_path(self, path):
    self.config_path = path
    return self.config_path

  def load_config(self):
    secrets_config_file = self.secrect_config_path
    config_file = self.config_path

    secret_file = open(secrets_config_file)
    secret = yaml.load(secret_file, Loader=SafeLoader)
    access_key = secret["storage"]["access_key"]
    secrets_key = secret["storage"]["secret_key"]
    endpoint = ""
    if "endpoint" in secret["storage"]:
      endpoint = ":".join([secret["storage"]["endpoint"], secret["storage"]["port"]])

    config_file = open(config_file)
    config = yaml.load(config_file, Loader=SafeLoader)
    buckets = config["storage"]["bucket"]["name"]

    config_file.close()
    secret_file.close()
    return buckets, endpoint, access_key, secrets_key


minio = MinioConnector()
