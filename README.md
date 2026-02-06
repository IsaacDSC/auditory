# auditory

Serviço de coleta e armazenamento de dados de auditoria com integração a bucket.

## Arquitetura

```
┌─────────────────────────────────────────────────────────────────┐
│                        Control Plane                            │
├─────────────────────────────────────────────────────────────────┤
│  HTTP Handlers              │  Background Tasks                 │
│  ───────────────            │  ─────────────────                │
│  POST /audit                │  • Backup (S3)                    │
│  POST /manual/backup        │  • Store (persist + cleanup)      │
│  POST /manual/store         │  • Idempotency clear              │
│  GET  /health               │                                   │
└──────────────┬──────────────┴───────────────┬───────────────────┘
               │                              │
               ▼                              ▼
┌──────────────────────────┐    ┌─────────────────────────────────┐
│    Local File Store      │    │         S3 / MinIO              │
│    (tmp/*.json)          │───▶│   audits/{key}/{date}.json      │
│    + Idempotency (mem)   │    │   audits/{date}.json (backup)   │
└──────────────────────────┘    └─────────────────────────────────┘
```

## Fluxo

1. **Recebe** evento de auditoria via `POST /audit`
2. **Valida** idempotência (evita duplicatas)
3. **Armazena** localmente em arquivo JSON agrupado por chave/data
4. **Sincroniza** periodicamente com S3 (backup + store)
5. **Limpa** dados locais antigos após persistência

## Estrutura

```
cmd/control-plane/     # Aplicação principal
├── main.go            # Entrypoint, setup de handlers e tasks
└── internal/
    ├── handle/        # HTTP handlers
    └── tasks/         # Background workers

internal/
├── audit/             # Modelo de dados (DataAudit)
├── backup/            # Serviços de backup e auditoria
├── cfg/               # Configuração da aplicação
└── store/             # Persistência (file + S3)
```


<!-- Melhorar a governança -->