kubectl apply --context aks-harbor-1-we1-d-main-admin -f ssh.yaml
k9s - port forward to openssh pod in default namesapce - port 2222

ssh -D 2223 linuxserver.io@localhost -p 2222
password: password

chromium --proxy-server="socks5://localhost:2223"

Registry: https://artifactorytest.cc.azd.cloud.allianz/ui/repos/tree/General/docker-registryman
Username: tu-registryman
Password: QzCH6Dz3EJ5c