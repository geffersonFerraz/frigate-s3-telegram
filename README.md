# Frigate-S3-Telegram

Frigate-S3-Telegram is a Go-based application that integrates Frigate, S3, and Telegram to manage and notify about security events. The application fetches events from Frigate, processes them, and sends notifications to a Telegram bot. It also stores event data in an S3 bucket and uses RabbitMQ for message queuing.

## Features

- Fetch events from Frigate
- Send event snapshots to Telegram
- Store event data in an S3 bucket
- Use RabbitMQ for message queuing
- Redis for caching event IDs

## Architecture

```mermaid
flowchart TD
    A(Frigate-S3-Telegram <br>Main loop) --> B[Get Frigate Events]
    B --> C{Return at least one event 'In progress'}
    C --> |No|A
    C --> |Yes|D{Event ID already exists in Redis?}
    D -->|Yes| A
    D -->|No| E[Add event ID to Redis]
    E --> RE[Send the snapshot to telegram]
    RE --> SE[Publish message to RabbitMQ]
    

    G(Rabbit Consumer) -->H[Get Frigate Event using ID]
    H -->I{Still in progress?}
    I -->|Yes|J[NACK message]
    J --> G
    I -->|No|K(Get MP4 file from Frigate)
    K -->L{Is greater than 50MB?}
    L -->|Yes|M(Send to S3 Bucket)
    M -->O(Send snapshot with presigned url to Telegram)
    L -->|No|N(Send to Telegram)
    N --> S
    O --> S(ACK Message)

```

<details>
  <summary>Screenshots</summary>
  
  ### Telegram
  

  ### Bucket
  
</details>