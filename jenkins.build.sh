#!/bin/bash

# install pnpm
npm install -g pnpm@9.12.2 --registry=https://registry.npmmirror.com/

# install (dev)dependencies
pnpm install --registry=https://registry.npmmirror.com/

# build dist
pnpm websrv:build

# remove original dist files
rm -rf /opt/shares/naotodoserver

# copy new dist files
cp -r apps/main/dist /opt/shares/naotodoserver

# bash into the node container
docker exec -it mynode /bin/bash

# start node
pm2 restart server

