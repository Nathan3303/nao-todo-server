# stop docker container mynode
docker stop mynode

# install pnpm
npm install -g pnpm --registry=https://registry.npmmirror.com

# install (dev)dependencies
pnpm install --registry=https://registry.npmmirror.com

# build dist files
pnpm run websrv build

# copy dist files into directory bundle
cp -r apps/web/dist/* apps/web/bundle/

# remove original dist files
rm -rf /opt/shares/naotodoserver

# copy dist files to folder /opt/shares(jenkins_share)
cp -r apps/web/bundle /opt/shares/naotodoserver

# start docker container mynode
docker start mynode

# bash into docker container mynode
docker exec -i mynode /bin/bash -c "pm2 status && cd /opt/shares/naotodoserver && npm install --registry=https://registry.npmmirror.com && pm2 start index.js"

# check pm2 status (optional)
#pm2 status

# move to folder /opt/shares/naotodoserver_t
#cd /opt/shares/naotodoserver_t || exit

# list files (optional)
#ls -l

# exit docker container mynode
#exit