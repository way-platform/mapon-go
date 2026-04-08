# Mapon Data Forwarding API Documentation

## Overview

The Data Forwarding API enables management of data endpoints for forwarding unit information to external systems. All endpoints require authentication via API key.

## Authentication

All requests require an API key passed as the `key` parameter:
- For GET requests: query string parameter `?key=...`
- For POST requests: form field `key=...`

## Available Endpoints

### List Forwarding Endpoints

**Method:** `GET https://mapon.com/api/v1/data_forward/list.json`

Retrieve all configured forwarding endpoints for the authenticated API key.

**Parameters:**
- `key` — API key (required)

**Response:**
Returns array of endpoint objects with:
- `id` — Endpoint ID (integer)
- `url` — Webhook URL
- `packs` — Array of pack IDs configured for this endpoint
- `unit_ids` — Array of unit IDs (empty = all units)
- Additional metadata (creation date, status, etc.)

---

### Create/Update Forwarding Endpoint

**Method:** `POST https://mapon.com/api/v1/data_forward/insert_update.json`

Create new or modify existing data forwarding endpoints.

**Parameters:**
- `key` — API key (required)
- `url` — Webhook URL to receive data (required if creating)
- `packs` — Comma-separated or array of pack IDs to forward (e.g., "1,3,5,26,55")
- `unit_ids` — Comma-separated or array of unit IDs (empty/omitted = all units)
- `id` — Endpoint ID (only when updating existing endpoint)

**Response:**
- `id` — Endpoint ID (integer)
- `url` — Confirmed webhook URL
- `packs` — Confirmed pack configuration

**Notes:**
- Cannot mix pack types within a single endpoint scope (e.g., cannot have both CAN and car packs)
- Empty `unit_ids` means the endpoint receives data for all units under this API key

---

### Delete Forwarding Endpoint

**Method:** `POST https://mapon.com/api/v1/data_forward/delete.json`

Remove an entire data forwarding endpoint and its associated configuration.

**Parameters:**
- `key` — API key (required)
- `id` — Endpoint ID to delete (required)

**Response:**
- Success indicator (typically `{"status":"ok"}`)

---

### Add Unit to Endpoint

**Method:** `POST https://mapon.com/api/v1/data_forward/add_unit_id.json`

Attach a single unit ID to an existing endpoint for data forwarding.

**Parameters:**
- `key` — API key (required)
- `id` — Endpoint ID (required)
- `unit_id` — Unit ID to add (required)

**Response:**
- Updated endpoint configuration

---

### Remove Unit from Endpoint

**Method:** `POST https://mapon.com/api/v1/data_forward/remove_unit_id.json`

Detach a single unit from an endpoint's forwarding configuration.

**Parameters:**
- `key` — API key (required)
- `id` — Endpoint ID (required)
- `unit_id` — Unit ID to remove (required)

**Response:**
- Updated endpoint configuration

---

### Purge Endpoint Queue

**Method:** `POST https://mapon.com/api/v1/data_forward/purge.json`

Clear pending data packets queued for transmission to an endpoint.

**Parameters:**
- `key` — API key (required)
- `id` — Endpoint ID (required)

**Response:**
- Purge status confirmation

---

### List Data Packs (Historical/Queued)

**Method:** `GET https://mapon.com/api/v1/data_forward/list_data_packs.json`

View queued or historical data packets associated with specific endpoints.

**Parameters:**
- `key` — API key (required)
- `id` — Endpoint ID (optional, filter by endpoint)
- Other filters as needed

**Response:**
- Array of pack objects with timestamps, payloads, and delivery status

---

## Integration with Push Connector

For push connector registration, use **insert_update** to create an endpoint with:
```
url=https://{connector-service-url}/webhook/organizations/{organization_id}/integrations/{integration_id}
packs=1,3,5,26,55
unit_ids=  # empty = all units for this API key
```

The returned `id` should be stored as an annotation on the Integration resource for later deregistration.

When deleting the integration, call **delete** with the stored endpoint `id`.

