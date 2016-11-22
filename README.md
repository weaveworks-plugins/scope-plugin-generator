# Scope Plugin Generator
The Scope Plugin Generator uses [Cookiecutter](https://github.com/audreyr/cookiecutter) to automatically generate a [Weave Scope](https://github.com/weaveworks/scope) plugin's basic structure in Go.

You can start developing your own Scope plugins in less then 5 minutes by just following the instructions in the next section.

## Plugin Bootstrapping
To generate your own plugin you will need Cookiecutter, you can install it following these [instructions](https://cookiecutter.readthedocs.io/en/latest/installation.html).

After you have installed Cookiecutter, you can create a basic Scope plugin structure using the following command: `cookiecutter https://github.com/weaveworks-plugins/scope-plugin-generator`.
You will be prompted to insert the following basic information about your plugin.

Mandatory attributes to generate a working plugin:
- `project_name` - project name.
- `plugin_id` - plugin id, for more information check Weave Scope [documentation](https://www.weave.works/documentation/scope-latest-plugins/#plugin-id).
- `plugin_name` - plugin name, this will be visible in the Scope UI.

Optional attributes:
- `plugin_description` - a brief description of the plugin.
- `maintainer_organization` - maintainer's organization name.
- `maintainer_email` - maintainer's email address.
- `docker_organization` - docker organization name.
- `docker_username` - docker username.
- `docker_email` - docker user's email address.
- `docker_repo` - docker repository name.

The maintainer and docker attributes are used for setting up the `circle.yml` file and to properly tag docker images.
They are not necessary to generate a working plugin, for simplicity you can use the default values.

The configuration file for the bootstrapping process is [cookiecutter.json](cookiecutter.json).

The plugin generator also creates the local git repository.

## What to expect

Run `cookiecutter --no-input https://github.com/weaveworks-plugins/scope-plugin-generator` to generate a plugin with the default values. After the generation process is finished, the directory tree will look as follows:

```
scope-awesome-plugin/
├── Dockerfile
├── LICENSE
├── Makefile
├── README.md
├── circle.yml
├── main.go
└── tools
```

- `Dockerfile` - used to build the docker image for the plugin.
- `Makefile` - provides commands to build, run, and clean.
- `README.md` - contains the name of the project and the project description.
- `circle.yml` - ready to use circleci configuration file.
- `main.go` - simple plugin skeleton in Go. It reports:
 - two values, one in the general section and the other in a plugin table).
 - two controls, by design only one at the time will appear.

- `tools/` - contains the [Weaveworks build-tools](https://github.com/weaveworks/build-tools). This is set up by [hooks/post\_gen\_project.sh](hooks/post_gen_project.sh), after the project is generated, using `git subtree`.

To build and run the generated plugin type:

```
cd scope-awesome-plugin/
make
```

**Note**: The default values are useful just as an example, you should **NOT** use the default values for your plugins, instead you _must_ insert meaningful values.
