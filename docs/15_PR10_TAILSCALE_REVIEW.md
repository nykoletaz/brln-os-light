# PR #10 - Documentacao de Ajustes (Tailscale App)

Este documento resume os ajustes que precisam ser feitos na PR `#10` antes do merge, para reduzir risco de regressao em funcionalidades existentes.

## Escopo analisado

- PR: `https://github.com/jvxis/brln-os-light/pull/10`
- Objetivo da PR: adicionar o app Tailscale na App Store (backend + UI).
- Arquivos principais afetados na PR:
  - `lightningos-light/internal/server/apps_tailscale.go`
  - `lightningos-light/internal/server/apps_registry.go`
  - `lightningos-light/internal/server/routes.go`
  - `lightningos-light/ui/src/pages/Tailscale.tsx`
  - `lightningos-light/ui/src/pages/AppStore.tsx`
  - `lightningos-light/ui/src/App.tsx`
  - `lightningos-light/ui/src/api.ts`
  - `lightningos-light/ui/src/i18n/en.json`
  - `lightningos-light/ui/src/i18n/pt-BR.json`

## Ajustes bloqueadores (antes de aprovar)

### 1) Instalar sem `curl | sh` como root

- Onde:
  - `lightningos-light/internal/server/apps_tailscale.go` (metodo `Install`, trecho com `curl -fsSL ... | sh`)
- Problema:
  - Executa script remoto sem pin de versao/checksum.
  - Aumenta risco de supply-chain e comportamento nao deterministico no host.
- Ajuste esperado:
  - Trocar por fluxo de instalacao mais controlado e auditavel:
    - pacote com versao pinada, ou
    - repositorio apt com chave + pinning + validacao de origem.
  - Retornar erro claro quando instalacao falhar.

### 2) Remover efeito colateral de endpoint GET

- Onde:
  - `lightningos-light/internal/server/routes.go` (`GET /api/apps/tailscale/auth-url`)
  - `lightningos-light/internal/server/apps_tailscale.go` (`handleTailscaleAuthURL`)
  - `lightningos-light/ui/src/pages/Tailscale.tsx` (polling `setInterval(loadInfo, 15000)`)
- Problema:
  - Endpoint `GET` dispara `tailscale up --accept-routes`, que altera estado de rede.
  - Com polling de 15s, essa acao pode ser repetida em loop.
- Ajuste esperado:
  - `GET /auth-url` deve ser somente leitura (sem side effect).
  - Criar endpoint `POST` explicito para iniciar login (`tailscale up`), chamado por acao do usuario.
  - Evitar polling agressivo para acao de login (usar acao manual + refresh de status).

### 3) Nao ignorar erros em pontos criticos

- Onde:
  - `lightningos-light/internal/server/apps_tailscale.go`
    - `Start()` (saida/erro de `tailscale up`)
    - `handleTailscaleAuthURL()` (saida/erro de `tailscale up`)
- Problema:
  - Erros sao descartados (`out, _ := ...`), podendo retornar sucesso falso na API/UI.
- Ajuste esperado:
  - Capturar e propagar erro com mensagem objetiva via `writeError(...)`.
  - Nao salvar auth URL quando o comando falhar.
  - Garantir consistencia de estado (UI nao mostrar fluxo como OK sem sucesso real).

## Ajustes recomendados (nao bloqueadores, mas importantes)

### 4) Cobertura minima de testes

- Onde:
  - `lightningos-light/internal/server/*_test.go`
  - `lightningos-light/ui/src/*` (se houver suite de testes frontend)
- Ajuste esperado:
  - Testes de handler para:
    - `GET /api/apps/tailscale/status`
    - `GET /api/apps/tailscale/auth-url` (somente leitura)
    - `POST /api/apps/tailscale/logout`
  - Validar cenarios de erro e de comando indisponivel.

### 5) Semantica de status na UI

- Onde:
  - `lightningos-light/ui/src/pages/Tailscale.tsx`
- Ajuste esperado:
  - Separar claramente:
    - status do service (`running/stopped`)
    - status de login (`NeedsLogin/Running`)
  - Mensagens de erro devem refletir erro real de backend.

## Criterio de aceite para merge

Aceitar a PR somente quando:

1. Instalacao nao depender de `curl | sh` sem controle de versao/origem.
2. Endpoint GET nao alterar estado (sem `tailscale up` em GET).
3. Erros de `tailscale up` nao forem ignorados.
4. Fluxo de login estiver em endpoint `POST` e UI sem reexecucao automatica indevida.
5. Testes minimos cobrirem os handlers novos.

## Texto curto sugerido para responder na PR

```txt
Obrigado pela contribuicao. A feature esta boa, mas para evitar regressao no host e na rede precisamos de alguns ajustes antes do merge:

1) Remover install via `curl | sh` (usar instalacao controlada/pinada).
2) `GET /api/apps/tailscale/auth-url` nao pode executar `tailscale up` (GET sem side effects).
3) Nao ignorar erros de `tailscale up` em Start/AuthURL handler.
4) Ajustar UI para fluxo de login por acao explicita (POST), sem polling que repete acao.
5) Adicionar cobertura minima de testes para os handlers novos.

Com esses pontos corrigidos, revalidamos para merge.
```
