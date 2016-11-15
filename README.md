# Scope Plugin Generator
The Scope Plugin Generator uses [Cookiecutter](https://github.com/audreyr/cookiecutter) to automatically generate a [Weave Scope](https://github.com/weaveworks/scope) plugin's basic structure in Go.

You can start developing your own Scope plugins in less then 5 minutes by just following the instructions in the next section.

## Plugin Bootstrapping
To generate your own plugin you will need Cookiecutter, you can install it following these [instructions](https://cookiecutter.readthedocs.io/en/latest/installation.html).

After you have installed Cookiecutter, you can create a basic Scope plugin structure using the following command: `cookiecutter https://github.com/weaveworks-plugins/scope-plugin-generator`.
You will be prompted to insert the following basic information about your plugin:

- `maintainer_organization` - maintainer's organization name.
- `maintainer_email` - maintainer's email address.
- `project_name` - project name.
- `project_root_dir_name` - name for the project root directory.
- `project_short_description` - project description.
- `version` - project version.
- `open_source_license` - project license.
- `golang_version` - golang version.
- `plugin_id` - plugin id, for more information check Weave Scope [documentation](https://www.weave.works/documentation/scope-latest-plugins/#plugin-id).
- `plugin_name` - plugin name, this will be visible in the Scope UI.
- `plugin_description` - a brief description of the plugin.
- `docker_organization` - docker organization name.
- `docker_username` - docker username.
- `docker_email` - docker user's email address.
- `docker_repo` - docker repository name.

The configuration file for the bootstrapping process is [cookiecutter.json](cookiecutter.json).
