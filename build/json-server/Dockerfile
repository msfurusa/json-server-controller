FROM node:20-alpine

RUN npm install -g json-server

ENV PATH="/usr/local/bin:${PATH}"

WORKDIR /data
EXPOSE 3000

ENTRYPOINT ["/usr/local/bin/json-server", "--watch", "/data/db.json", "--host", "0.0.0.0", "--port", "3000"]
