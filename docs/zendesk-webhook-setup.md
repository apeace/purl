# Zendesk Webhook Setup

Purl receives live updates from Zendesk via webhook event subscriptions. This
keeps tickets, comments, customers, and agents in sync without waiting for a
manual `pull-zendesk` run.

---

## Prerequisites

Each organization needs its Zendesk credentials stored in the database:

- `zendesk_subdomain` — e.g. `acme` for `acme.zendesk.com`
- `zendesk_email` — the email of the Zendesk admin/agent used for API access
- `zendesk_api_key` — the Zendesk API token (not the account password)
- `zendesk_webhook_secret` — generated automatically by `create-org`; used to
  verify that incoming requests actually come from Zendesk

If you need to retrieve the webhook secret for an existing org:

```sql
SELECT zendesk_webhook_secret FROM organizations WHERE slug = 'your-org-slug';
```

---

## Step 1 — Create the webhook in Zendesk Admin Center

1. Open **Admin Center** → **Apps and Integrations** → **Webhooks**
2. Click **Create webhook**
3. Fill in the form:

   | Field | Value |
   |---|---|
   | **Name** | `Purl Live Sync` |
   | **Endpoint URL** | `https://<your-purl-host>/webhooks/zendesk/<org-slug>` |
   | **Request method** | POST |
   | **Request format** | JSON |
   | **Authentication** | None |

4. Under **Signing**, enable **Sign webhook** and copy the generated signing
   secret into your org row (`zendesk_webhook_secret`). Alternatively, use the
   value Purl generated at org creation time and paste it into the Zendesk form.
5. Click **Create webhook**

---

## Step 2 — Subscribe to events

After creating the webhook, click **Subscriptions** on the webhook detail page
and add the following event types:

| Event type | What it covers |
|---|---|
| `zen:event-type:ticket.created` | New ticket opened |
| `zen:event-type:ticket.updated` | Title, description, status, or assignee changed |
| `zen:event-type:ticket.deleted` | Ticket deleted or merged (source ticket) |
| `zen:event-type:comment.created` | New comment (public reply or internal note) |
| `zen:event-type:comment.updated` | Comment body redacted by an agent |
| `zen:event-type:user.created` | New end-user or agent account created |
| `zen:event-type:user.updated` | Name, email, or role changed |

> **Note:** `user.created` and `user.updated` fire for all Zendesk user types
> (end-users, agents, admins). Purl routes them to `customers` or `agents`
> based on the `role` field in the payload.

---

## Step 3 — Test the webhook

1. On the webhook detail page, click **Test webhook**
2. Choose any event type and click **Send test**
3. Purl responds with `204 No Content` on success

If you receive a `401 Unauthorized`, the signing secret does not match what is
stored in the database.

---

## How Purl handles each event

### Ticket created / updated
- Upserts the ticket row (keyed on `zendesk_ticket_id`)
- If the requester is not in the database, Purl fetches them from the Zendesk
  REST API automatically
- On **status change** (or new ticket), moves the ticket to the matching column
  in the org's default Kanban board. If the new status has no matching column,
  the ticket is removed from the default board. Position is appended at the end
  of the column.

### Ticket deleted
- Deletes the ticket. Cascades to `ticket_comments` and `board_tickets`.

### Comment created
- Upserts the comment row (keyed on `zendesk_comment_id` within the ticket)
- Maps `via.channel` + `public` flag to our `comment_channel` enum:

  | Zendesk via.channel | public | Stored as |
  |---|---|---|
  | `email` | true | `email` |
  | `sms`, `native_messaging`, `whatsapp` | true | `sms` |
  | `voice`, `phone` | true | `voice` |
  | anything else | true | `web` |
  | any | **false** | `internal` |

- If the comment author is not in the database, Purl fetches them from Zendesk

### Comment updated
- Updates `body` on the matching comment (handles agent redaction)

### User created / updated
- `end-user` → upserted into `customers` and `customer_emails`
- `agent` / `admin` → upserted into `agents`

---

## Signature verification

Zendesk signs every request with HMAC-SHA256:

```
Signature = base64( HMAC-SHA256( signing_secret, timestamp + body ) )
```

The signature is in the `X-Zendesk-Webhook-Signature` header; the timestamp is
in `X-Zendesk-Webhook-Signature-Timestamp`. Purl rejects any request where the
computed signature does not match.

---

## After initial setup

Run `pull-zendesk` **once** after configuring credentials to do the initial
full sync. After that, the webhook keeps everything current in real time.

```bash
DATABASE_URL=... go run ./cmd/pull-zendesk <org-slug>
```
