# stop docker container mynode, mymongo and mynginx
docker stop mynode mymongo mynginx

# install pnpm, and install (dev)dependencies
npm install -g pnpm --registry=https://registry.npmjs.org
pnpm install --registry=https://registry.npmjs.org

# build dist files, and copy dist files into directory bundle
pnpm websrv build
cp -r apps/web/dist/* apps/web/bundle/

# remove original dist files, and copy dist files to folder /opt/shares(jenkins_share)
rm -rf /opt/shares/naotodoserver
cp -r apps/web/bundle /opt/shares/naotodoserver

# start docker container mynginx, mynode and mymongo
docker start mynginx mynode mymongo

# bash into docker container mynode
docker exec -i mynode /bin/bash -c "pm2 status && cd /opt/shares/naotodoserver && npm install && npm run start"