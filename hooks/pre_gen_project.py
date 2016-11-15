import re
import sys

NAME_REGEX = r'^[A-Za-z0-9]+([-][A-Za-z0-9]+)*$'

plugin_id = '{{ cookiecutter.plugin_id }}'
docker_repo = '{{ cookiecutter.docker_repo }}'

if not re.match(NAME_REGEX, plugin_id):
    print('ERROR: %s is not a valid scope plugin id!' % plugin_id)
    print('For more information read https://www.weave.works/documentation/scope-latest-plugins/#plugin-id')

    # exits with status 1 to indicate failure
    sys.exit(1)

if not re.match(NAME_REGEX, docker_repo):
    print('ERROR: %s is not a valid docker repository name!' % docker_repo)

    # exits with status 1 to indicate failure
    sys.exit(1)
