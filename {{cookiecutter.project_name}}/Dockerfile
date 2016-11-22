FROM alpine:3.3
MAINTAINER {{cookiecutter.maintainer_organization}} <{{cookiecutter.maintainer_email}}>
LABEL works.weave.role=system
COPY ./{{cookiecutter.plugin_id}} /usr/bin/{{cookiecutter.plugin_id}}
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
ENTRYPOINT ["/usr/bin/{{cookiecutter.plugin_id}}"]
