# FILE: docs/04_INSTALLER_SPEC.md

# Installer Spec (v0.1)

## Objetivo
Instalar e configurar:
- PostgreSQL
- LND (binário)
- LightningOS Manager (binário)
- serviços systemd
- preparar configs padrão para Bitcoin remoto BRLN
- subir UI em 127.0.0.1:8443

## Requisitos
- Ubuntu Server (22.04/24.04)
- acesso root/sudo
- pacotes:
  - postgresql
  - smartmontools
  - ufw (opcional)
  - curl, jq, ca-certificates

## Etapas do instalador (script)
1) Verificar OS + arquitetura
2) Criar usuários:
- useradd --system --home /home/lnd --shell /usr/sbin/nologin lnd
- useradd --system --home /var/lib/lightningos --shell /usr/sbin/nologin lightningos
- criar /data/lnd e ajustar owner para lnd
- adicionar `lightningos` ao grupo `lnd` para leitura de tls/macaroon

3) Instalar pacotes:
- apt update
- apt install -y postgresql smartmontools curl jq ca-certificates

4) Configurar PostgreSQL:
- criar role `lndpg` com senha aleatória
- criar db `lnd`
- grants mínimos
- salvar DSN em `/etc/lightningos/secrets.env` (chmod 660, root:lightningos)

5) Instalar LND binário:
- baixar release oficial (URL param)
- validar checksum se disponível
- colocar em /usr/local/bin/lnd e lncli
- chmod +x

6) Criar /data/lnd/lnd.conf (template) com:
- mainnet only
- bitcoind remoto default (host+zmq)
- placeholders rpcuser/rpcpass (vazios inicialmente)
- postgres backend DSN referenciando envfile ou include

7) Instalar LightningOS Manager:
- /opt/lightningos/manager/lightningos-manager
- UI build em /opt/lightningos/ui/ (ou embutida no binário)

8) TLS para UI:
- gerar self-signed cert em /etc/lightningos/tls/
- manager usa esses arquivos

9) systemd unit files:
- /etc/systemd/system/lnd.service
- /etc/systemd/system/lightningos-manager.service

10) Enable e start:
- systemctl daemon-reload
- systemctl enable --now postgresql
- systemctl enable --now lnd
- systemctl enable --now lightningos-manager

## Resultado esperado
- UI disponível em https://127.0.0.1:8443
- Wizard guiará usuário a inserir rpcuser/rpcpass e criar/importar wallet

## Restrições
- não abrir portas WAN
- não armazenar seed phrase

