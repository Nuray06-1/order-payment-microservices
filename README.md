# 🌟 README — VERSION PRO

```markdown
# 🚀 Order & Payment Microservices System (Go)

## 🧩 Overview

This project implements a **distributed microservices system** consisting of two independent services:

- 🛒 **Order Service** — manages customer orders
- 💳 **Payment Service** — processes and validates payments

The system is built using **Go (Golang)** and follows **Clean Architecture principles**, ensuring scalability, maintainability, and clear separation of concerns.

---

## 🏗️ Architecture

Each service follows a layered Clean Architecture:

```

Transport (HTTP)
↓
Use Case (Business Logic)
↓
Domain (Entities)
↓
Repository (Persistence)

```

### 🔹 Key Design Principles

- ✅ Separation of Concerns  
- ✅ Dependency Inversion  
- ✅ Interface-driven design  
- ✅ Thin HTTP handlers  
- ✅ Independent services  

---

## 🔗 Microservices Design

| Feature | Implementation |
|--------|------|
| Services | Order + Payment |
| Communication | REST (HTTP) |
| Data Ownership | Separate databases |
| Coupling | Loose |
| Deployment | Independent |

👉 Each service owns its own data and logic — no shared models or databases.

---

## 🌐 Service Interaction

```

Client
↓
Order Service → Payment Service
↓                 ↓
Order DB        Payment DB

````

- Order Service sends a request to Payment Service  
- Payment Service validates and responds  
- Order status is updated accordingly  

---

## 🛠️ Tech Stack

- **Language:** Go (Golang)
- **Framework:** Gin (HTTP)
- **Database:** PostgreSQL
- **Architecture:** Clean Architecture
- **Communication:** REST API
- **ID generation:** UUID

---

## 💾 Database Design

Each service has its own PostgreSQL database:

| Service | Database |
|--------|--------|
| Order Service | `orderdb` |
| Payment Service | `paymentdb` |

👉 This ensures **data isolation and service independence**

---

## 📌 API Endpoints

### 🛒 Order Service

| Method | Endpoint | Description |
|------|--------|------------|
| POST | `/orders` | Create new order |
| GET | `/orders/{id}` | Get order details |
| PATCH | `/orders/{id}/cancel` | Cancel order |

---

### 💳 Payment Service

| Method | Endpoint | Description |
|------|--------|------------|
| POST | `/payments` | Process payment |
| GET | `/payments/{order_id}` | Get payment status |

---

## 🧠 Business Rules

### 💰 Financial Accuracy
- Amount is stored as `int64`
- Avoids floating-point precision issues

---

### 📦 Order Rules

- Amount must be **> 0**
- Status flow:
  - `Pending → Paid`
  - `Pending → Failed`
- ❌ **Paid orders cannot be cancelled**

---

### 💳 Payment Rules

- If amount > **100000** → ❌ Declined
- Otherwise → ✅ Authorized

---

## ⚠️ Failure Handling

If Payment Service is unavailable:

- ⛔ Order Service returns **503 Service Unavailable**
- ⏱️ Timeout prevents hanging requests
- 📉 Order is marked as **Failed**

---

## 🔒 Reliability Features

- ⏱️ HTTP client timeout (max 2 seconds)
- 🔁 Safe error handling
- 🧱 No shared state between services
- 🧬 UUID-based unique identifiers

---

## 🧪 Example Requests

### ➕ Create Order

```bash
curl -X POST http://localhost:8081/orders \
-H "Content-Type: application/json" \
-d '{
  "customer_id": "123",
  "item_name": "Coffee",
  "amount": 5000
}'
````

---

### 🔍 Get Order

```bash
curl http://localhost:8081/orders/{id}
```

---

### ❌ Cancel Order

```bash
curl -X PATCH http://localhost:8081/orders/{id}/cancel
```

---

### 💳 Get Payment

```bash
curl http://localhost:8083/payments/{order_id}
```

---

## 🎯 Design Highlights

✔ Clean separation between layers
✔ Independent microservices
✔ Real database persistence
✔ Proper business rule enforcement
✔ Resilient communication via timeouts

---

## 📊 Conclusion

This project demonstrates a **real-world microservices architecture** with:

* Clean Architecture implementation
* Strong separation of concerns
* Reliable inter-service communication
* Production-level design patterns

👉 The system is scalable, testable, and ready for extension.

---

## 💡 Author

Developed as part of **Advanced Programming 2 (AP2)** course.

---

```

---