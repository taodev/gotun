FROM ubuntu

COPY ./gotun /apps/

WORKDIR /apps

ENTRYPOINT [ "/apps/gotun", "run", "-c", "/etc/gotun/config.yaml" ]