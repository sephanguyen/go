#!/bin/bash

# This script installs the required tools to run backend cluster locally.
# Only Linux/Ubuntu is supported. For other OSes, please install them on your own.
# Only bash is fully supported. For other shells (zsh, fish, etc..), you will need
# to add $HOME/.manabie/bin to your PATH manually.

set -euo pipefail

MANABIE_HOME="${MANABIE_HOME:-"$HOME/.manabie"}"

# shellcheck source=../scripts/log.sh
source ./scripts/log.sh

setup_profile() {
  # Require any of these 2 to be present in PATH
  # shellcheck disable=SC2088
  local require_path="~/.manabie/bin"
  local require_path2="$MANABIE_HOME/bin"
  local pfound=false
  if echo "$PATH" | grep -F "$require_path" >/dev/null 2>&1; then
    pfound=true
  fi
  if echo "$PATH" | grep -F "$require_path2" >/dev/null 2>&1; then
    pfound=true
  fi
  if [[ "$pfound" != "true" ]]; then
      logerror "Please add $MANABIE_HOME/bin to your PATH. You can usually do this by adding:
    export PATH=\$PATH:$MANABIE_HOME/bin
to your \$HOME/.profile or \$HOME/.bashrc"
    return 1
  fi
}

# Check if the program is installed by this script.
# If not, return an error
check_managed() {
  local program=$1
  local path=$2
  local required_version=$3
  if [[ "$binpath" != "$MANABIE_HOME"* ]]; then
    logerror "$program is installed at $path, which is not managed by this script. \
Please either install $program $required_version manually, or uninstall the existing version and run this script again."
    return 1
  fi
}

check_exist() {
  local program=$1
  if ! binpath=$(command -v "$program"); then
    if [ -f "$MANABIE_HOME/bin/$program" ]; then
      logerror "$program exists at $MANABIE_HOME/bin/$program, but is not in your PATH. \
Please add $MANABIE_HOME/bin to your PATH. You can usually do this by adding:
    export PATH=\$PATH:$MANABIE_HOME/bin
to your \$HOME/.profile or \$HOME/.bashrc"
    else
      logerror "Failed to install $program ($MANABIE_HOME/bin/$program not found)"
    fi
    return 1
  fi
}

download_install() {
  binname="$1"
  sourceurl="$2"
  mkdir -p "$MANABIE_HOME/bin/"
  curl -fLo "$MANABIE_HOME/bin/${binname}" "${sourceurl}"
  chmod +x "$MANABIE_HOME/bin/${binname}"
  check_exist "${binname}"
}

# noversion is the fallback version string when the command
# to get the version of a binary fails, so that this script
# can continue to (re)install the binary
noversion() {
  echo "failed-to-get-version"
}

get_required_version() {
  binname="$1"
  cat "${BASH_SOURCE%/*}/versions/$1" || noversion
}

setup_skaffold() {
  local required_version="$(get_required_version skaffold)"
  local upgrade=false
  if ! binpath=$(command -v skaffold); then
    upgrade=true
  else
    existing_version=$(skaffold version -o "{{.Version}}" || noversion)
    if [ "$existing_version" != "$required_version" ]; then
      check_managed "skaffold" "$binpath" "$required_version"
      upgrade=true
    fi
  fi

  if [ "$upgrade" != "true" ]; then
    logdebug "skaffold $required_version already installed at $binpath"
  else
    echo "Installing skaffold $required_version"
    download_install "skaffold" "https://github.com/GoogleContainerTools/skaffold/releases/download/${required_version}/skaffold-linux-amd64"

    loginfo "skaffold $required_version has been installed at $binpath"
  fi
}

setup_skaffoldv2() {
  local required_version="$(get_required_version skaffoldv2)"
  local upgrade=false
  if ! binpath=$(command -v skaffoldv2); then
    upgrade=true
  else
    existing_version=$(skaffoldv2 version -o "{{.Version}}" || noversion)
    if [ "$existing_version" != "$required_version" ]; then
      check_managed "skaffoldv2" "$binpath" "$required_version"
      upgrade=true
    fi
  fi

  if [ "$upgrade" != "true" ]; then
    logdebug "skaffoldv2 $required_version already installed at $binpath"
  else
    echo "Installing skaffoldv2 $required_version"
    download_install "skaffoldv2" "https://github.com/GoogleContainerTools/skaffold/releases/download/${required_version}/skaffold-linux-amd64"

    loginfo "skaffoldv2 $required_version has been installed at $binpath"
  fi
}

setup_kind() {
  get_kind_version() {
    read -a rawstring <<<"$(kind version)"
    echo "${rawstring[1]}"
  }
  local required_version="$(get_required_version kind)"
  local upgrade=false
  if ! binpath=$(command -v kind); then
    upgrade=true
  else
    existing_version=$(get_kind_version || noversion)
    if [ "$existing_version" != "$required_version" ]; then
      check_managed "kind" "$binpath" "$required_version"
      upgrade=true
    fi
  fi

  if [ "$upgrade" != true ]; then
    logdebug "kind $required_version already installed at $binpath"
  else
    echo "Installing kind $required_version"
    download_install "kind" "https://github.com/kubernetes-sigs/kind/releases/download/${required_version}/kind-linux-amd64"
    loginfo "kind $required_version has been installed at $binpath"
  fi
}

setup_inotify() {
  inotify__max_user_watches=$(sysctl -n fs.inotify.max_user_watches)
  if [ "$inotify__max_user_watches" -lt "524288" ]; then
    echo "fs.inotify.max_user_watches is too low ($inotify__max_user_watches). Setting it to 524288."
    echo "To persist this change, add \"fs.inotify.max_user_watches=524288\" to /etc/sysctl.conf"
    sudo sysctl fs.inotify.max_user_watches=524288
  fi

  inotify__max_user_instances=$(sysctl -n fs.inotify.max_user_instances)
  if [ "$inotify__max_user_instances" -lt "512" ]; then
    echo "fs.inotify.max_user_instances is too low ($inotify__max_user_instances). Setting it to 512."
    echo "To persist this change, add \"fs.inotify.max_user_instances=512\" to /etc/sysctl.conf"
    sudo sysctl fs.inotify.max_user_instances=512
  fi
}

setup_jq() {
  local required_version="$(get_required_version jq)"
  local upgrade=false
  if ! binpath=$(command -v jq); then
    upgrade=true
  else
    existing_version=$(jq --version | cut -d"-" -f2 || noversion)
    if [ "$existing_version" != "$required_version" ]; then
      check_managed "jq" "$binpath" "$required_version"
      upgrade=true
    fi
  fi

  if [ "$upgrade" != true ]; then
    logdebug "jq $required_version already installed at $binpath"
  else
    echo "Installing jq $required_version"
    download_install "jq" "https://github.com/stedolan/jq/releases/download/jq-${required_version}/jq-linux64"
    loginfo "jq $required_version has been installed at $binpath"
  fi
}

setup_yq() {
  local required_version="$(get_required_version yq)"
  local upgrade=false
  if ! binpath=$(command -v yq); then
    upgrade=true
  else
    existing_version=v$(yq --version | cut -d' ' -f4 || noversion)
    if [ "$existing_version" != "$required_version" ]; then
      check_managed "yq" "$binpath" "$required_version"
      upgrade=true
    fi
  fi

  if [ "$upgrade" != true ]; then
    logdebug "yq $required_version already installed at $binpath"
  else
    echo "Installing yq $required_version"
    download_install "yq" "https://github.com/mikefarah/yq/releases/download/${required_version}/yq_linux_amd64"
    loginfo "yq $required_version has been installed at $binpath"
  fi
}

os=$(uname)
if [ "$os" != "Linux" ]; then
  logwarn "Unsupported OS ($os). Please install these tools on your own:
  - skaffold $(get_required_version skaffold) from https://github.com/GoogleContainerTools/skaffold
  - skaffoldv2 $(get_required_version skaffoldv2) from https://github.com/GoogleContainerTools/skaffold
  - kind $(get_required_version kind) from https://github.com/kubernetes-sigs/kind
  - adjust inotify: https://kind.sigs.k8s.io/docs/user/known-issues/#pod-errors-due-to-too-many-open-files
  - jq $(get_required_version jq) from https://github.com/stedolan/jq
  - yq $(get_required_version yq) from https://github.com/mikefarah/yq"
  exit 0
fi

cmdname=${0##*/}
fullcmdname="./deployments/${cmdname}"
function usage() {
  cat <<EOF
Ensure that the necessary tools to deploy backend exists and have the correct versions.
Otherwise, download and install them.

Options:
  ${cmdname} [options]
    -h, --help: Print help.
EOF
}

# process arguments
while [[ $# -gt 0 ]]; do
  case "$1" in
    -h | --help)
      usage
      exit 0
      ;;
    *)
      logfatal "Unrecognized argument: $1
See \"${fullcmdname} --help\" for more information."
  esac
done

setup_profile
setup_skaffold
setup_skaffoldv2
setup_kind
setup_inotify
setup_jq
setup_yq
