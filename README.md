# CodeStash Server

CodeStash のバックエンド (Go) を ECS/Fargate で動かすためのコードと Terraform 定義です。フロントは Vercel でホスティング予定、DB は Supabase への移行を進行中です。

## 現状の構成
- **アプリ**: Go サーバ (`cmd/...`)。コンテナは ECR に push して ECS/Fargate で稼働予定。
- **インフラ (Terraform)**: `terraform/` で VPC/ALB/ECS/ECR/WAF を管理（1AZ・1NAT のシンプル構成）。HTTPS/ACM は未設定。
- **DB**: Supabase を利用予定。ローカル開発時のみ `docker-compose` で Postgres + migration を起動可能。

## ローカル開発
```bash
# サーバ起動（例）
go run ./cmd/...
```

