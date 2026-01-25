# SMAP Consumer Service

Flow diagrams and architecture for Consumer Service.

## Table of Contents
- [Email Sending Flow](#email-sending-flow)

---

## Email Sending Flow

Email sending flow from API request to email delivery.

```mermaid
sequenceDiagram
    actor User
    participant API as API Server
    participant AuthUC as Auth UseCase
    participant Producer as Email Producer
    participant RabbitMQ as RabbitMQ
    participant Consumer as Consumer Service
    participant SMTP_UC as SMTP UseCase
    participant SMTP as SMTP Server
    
    User->>API: POST /identity/authentication/send-otp
    API->>AuthUC: SendOTP(email, password)
    
    Note over AuthUC: Validate user & password
    Note over AuthUC: Generate OTP
    Note over AuthUC: Update user.OTP
    
    AuthUC->>Producer: PublishSendEmail(EmailData)
    
    Producer->>Producer: json.Marshal(message)
    Producer->>RabbitMQ: Publish to Exchange<br/>"smtp_send_email_exc"
    
    Note over RabbitMQ: Route to Queue<br/>"smtp_send_email"
    
    AuthUC-->>API: Success
    API-->>User: 200 OK
    
    Note over Consumer: Listening on queue...
    
    RabbitMQ->>Consumer: Deliver Message
    Consumer->>Consumer: json.Unmarshal(message)
    Consumer->>SMTP_UC: SendEmail(EmailData)
    
    SMTP_UC->>SMTP: DialAndSend(message)
    SMTP-->>SMTP_UC: Success
    SMTP_UC-->>Consumer: Success
    
    Consumer->>RabbitMQ: Ack(message)
    
    Note over User: Email received
```

**Key Points:**
- Asynchronous email sending decoupled from API response
- RabbitMQ ensures message delivery reliability
- Consumer processes messages in background
- SMTP UseCase handles actual email delivery
- Message acknowledgment after successful send

---

*Last updated: November 20, 2025*

