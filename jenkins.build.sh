# install pnpm
npm install -g pnpm --registry=https://registry.npmmirror.com

# install (dev)dependencies
pnpm install --registry=https://registry.npmmirror.com

# build dist files
pnpm websrv build

# copy certs/ into dist/
cp -r apps/web/certs apps/web/dist

# remove original dist files
rm -rf /opt/shares/naotodoserver_t

# copy dist files to folder /opt/shares(jenkins_share)
cp -r apps/web/dist /opt/shares/naotodoserver_t