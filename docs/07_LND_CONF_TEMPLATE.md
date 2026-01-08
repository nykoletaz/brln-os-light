# FILE: docs/07_LND_CONF_TEMPLATE.md

# lnd.conf Template (mainnet + remote bitcoind)

# Core
[Application Options]
debuglevel=info
maxpendingchannels=10

# Bitcoin (mainnet only)
[Bitcoin]
bitcoin.active=1
bitcoin.mainnet=1
bitcoin.node=bitcoind

[Bitcoind]
bitcoind.rpchost=bitcoin.br-ln.com:8085
bitcoind.rpcuser={{BITCOIN_RPC_USER}}
bitcoind.rpcpass={{BITCOIN_RPC_PASS}}
bitcoind.zmqpubrawblock=tcp://bitcoin.br-ln.com:28332
bitcoind.zmqpubrawtx=tcp://bitcoin.br-ln.com:28333

# RPC (local only)
[RPC]
rpclisten=127.0.0.1:10009
restlisten=127.0.0.1:8080

# DB backend: postgres
[db]
db.backend=postgres
# A DSN deve vir de uma fonte segura.
# Opção A: inserir diretamente com perms 600 no lnd.conf (menos ideal)
# Opção B: gerar arquivo /data/lnd/lnd.postgres.conf root-only e include (se suportado)
# Opção C: gerenciar via flag/env no service (preferível)
#
# Exemplo (se suportado):
# db.postgres.dsn=postgres://lndpg:...@127.0.0.1:5432/lnd?sslmode=disable

