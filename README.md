hld - https://drive.google.com/file/d/1V7jANaH95C4ZMsg8-DaXNQg5gPrd2475/view?usp=drive_link

Evently | Go, Gin, MongoDB, PostgreSQL, Redis, Kafka 

• Engineered a scalable microservices-based backend system, enabling seamless event browsing, ticket
booking/cancellation, and real-time analytics.

• Built a gateway service for authentication, routing, and rate limiting, capable of handling 100+ requests/sec.

• Implemented Redis caching to deliver low-latency access to frequently requested event data, estimated to reduce db
calls.

• Designed idempotency checks using Redis to prevent duplicate bookings during high concurrency.

• Integrated Kafka (Pub/Sub) with consumer groups to manage booking traffic surges and enable parallel processing
and potentially making the system faster under high load.

