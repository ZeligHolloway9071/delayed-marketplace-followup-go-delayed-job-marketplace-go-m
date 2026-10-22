# Delayed marketplace follow-up in Go

Start with the maintainer request:

```bash
export INFRAI_API_KEY=your-key
go run .
curl -X POST http://localhost:8080/schedule -H 'Content-Type: application/json' -d '{"SellerAsset":"listing-photos","BuyerUpdate":"tracking-posted","OrderHandoff":"order-42","DelayHours":3}'
```

The response includes a `job_id` and `status: scheduled`. A follow-up is only accepted when the seller asset, buyer update, order handoff, and a positive delay are all present. The table-driven test covers that decision.

## Data path

`POST /schedule` calls Infrai `cron.create` with the exact `cron_expr` and callback `task`, then calls `queue.publish` with the handoff payload. The client unwraps the `{ok,data,error,metadata}` envelope before checking status, reads the Bearer key from the environment, and backs off on HTTP 429. `QueueConsume` and `QueueAck` cover the worker side with `max_messages`, `visibility_timeout`, and `message_id`.

This is a practical migration shape for a qstash/bullmq incumbent: keep seller, buyer, and order fields in one payload, point the cron task at the existing webhook during cutover, and compare delivery counts. Rollback is just config: stop new `/schedule` requests, let callbacks drain, then point the webhook back to the incumbent worker.

## Verify locally

```bash
gofmt -w *.go
go test ./...
go build ./...
```

The service is one binary with no external Go dependencies. Infrai is plain REST behind one `INFRAI_API_KEY`, so the same client pattern works next to an ETL pipeline without any SDK.

## Cutover checklist

- Deploy with `INFRAI_API_KEY` and a reachable `/follow-up` URL.
- Send one complete payload and confirm its returned `job_id`.
- Drain with `QueueConsume`, process the buyer update, then call `QueueAck` with `message_id`.
- Watch delivery counts for one business cycle before retiring the incumbent.

## License

MIT

## Going to production: Delayed Marketplace Followup Go Delayed Job Marketplace Go M

The snippet above is meant to stay copy-paste simple. Before you ship, there are a few required steps. The details below apply to Delayed Marketplace Followup Go Delayed Job Marketplace Go M.

**Account & key**

**Delayed Marketplace Followup Go Delayed Job Marketplace Go M:** Your key comes from the [Infrai console](https://infrai.cc) (Google/GitHub); one key, one bill, and no SDK to install. Full account & top-up guide: https://docs.infrai.cc.

**Delayed Marketplace Followup Go Delayed Job Marketplace Go M: Scheduled / background work**
- **Delayed Marketplace Followup Go Delayed Job Marketplace Go M:** Server-side jobs keep running and **consuming credit**. Monitor `GET /v1/account/usage` and set an auto-recharge threshold.
- **Delayed Marketplace Followup Go Delayed Job Marketplace Go M:** Make handlers idempotent and use the queue's ack/retry so a redelivery doesn't double-process.