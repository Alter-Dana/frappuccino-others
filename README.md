# Frappuccino ☕️

> A PostgreSQL-powered backend for managing coffee shop operations.

![ERD](/ERD.jpeg)

---

## 🚀 Learning Objectives

- SQL & PostgreSQL
- CRUD operations with real data
- ERD (Entity-Relationship Diagrams)
- Full-text search and aggregations
- Containerization with Docker

---

## 🗂 Technologies Used

- Go (Standard Library only, PostgreSQL driver allowed)
- PostgreSQL
- Docker + Docker Compose
- SQL (DDL + DML)
- RESTful API principles

---

## 📌 How to Run

Make sure Docker is installed. From the root directory:

```bash
docker compose up
```

> The API will be available at **localhost:8080**

Database credentials:
- **Host**: db  
- **Port**: 5432  
- **User**: latte  
- **Password**: latte  
- **Database**: frappuccino  

---

## 🧱 Project Structure

- `init.sql`: Schema definitions and mock data
- `Dockerfile` / `docker-compose.yml`: Container setup
- `handlers/`: HTTP route implementations
- `repository/`: Data Access Layer with SQL queries

---

## 📄 Mandatory Endpoints

### ☕ Orders
- `POST /orders` — Create
- `GET /orders` — Read all
- `GET /orders/{id}` — Read by ID
- `PUT /orders/{id}` — Update
- `DELETE /orders/{id}` — Delete
- `POST /orders/{id}/close` — Close order

Request body:
```json
{
  "customer_name": "John Doe",
  "items": [
    {
      "menu_id": 1,
      "quantity": 2
    }
  ]
}
```

### 🍰 Menu
- `POST /menu`
- `GET /menu`
- `GET /menu/{id}`
- `PUT /menu/{id}`
- `DELETE /menu/{id}`

Request body:
```json
{
    "name": "Chocolate Muffin",
    "description": "Freshly baked muffin with strawberries",
    "price": 2.00,
    "inventory": [
      {
        "inventory_id": 2,
        "quantity": 5
      },
      {
        "inventory_id": 10,
        "quantity": 1
      },
      {
        "inventory_id": 1,
        "quantity": 2
      }
    ]
  }
```

### 🧂 Inventory
- `POST /inventory`
- `GET /inventory`
- `GET /inventory/{id}`
- `PUT /inventory/{id}`
- `DELETE /inventory/{id}`

Request body:
```json
{
    "name": "banana",
    "quantity": 100,
    "unit": "units",
    "categories": [
        "Fruit",
        "Sweetener"
    ]
}
```

---

## 📊 Reports & Aggregations

### 1. Number of Ordered Items  
`GET /orders/numberOfOrderedItems?startDate=YYYY-MM-DD&endDate=YYYY-MM-DD`
Response example:
```json
GET orders/numberOfOrderedItems?startDate=01.01.2025&endDate=30.01.2025
HTTP/1.1 200 OK
Content-Type: application/json

{
    "Barista Special": 11,
    "Blueberry Muffin": 26,
    "Caffe Latte": 30,
    "Caramel Macchiato": 13,
    "Chai Tea Latte": 5,
    "Chocolate Frappe": 10,
    "Double Espresso": 18,
    "Espresso": 17,
    "Ice Latte": 26,
    "Matcha Latte": 20,
    "Raspberry Muffin": 15,
    "Strawberry Muffin": 12,
    "Vanilla Cappuccino": 8
}
```

### 2. Full Text Search  
`GET /reports/search?q=...&filter=menu,orders&minPrice=10`
Response example:
```json
GET /reports/search?q=muffin&filter=menu
HTTP/1.1 200 OK
Content-Type: application/json

{
    "menu_items": [
        {
            "id": "1",
            "name": "Blueberry Muffin",
            "description": "Freshly baked muffin with blueberries",
            "price": 2,
            "relevance": 0.66871977
        },
        {
            "id": "2",
            "name": "Raspberry Muffin",
            "description": "Muffin with fresh raspberries",
            "price": 2,
            "relevance": 0.66871977
        },
        {
            "id": "3",
            "name": "Strawberry Muffin",
            "description": "Freshly baked muffin with strawberries",
            "price": 2,
            "relevance": 0.66871977
        }
    ],
    "total_matches": 3
}
```

### 3. Ordered Items by Period  
`GET /reports/orderedItemsByPeriod?period=month&year=2023`  
`GET /reports/orderedItemsByPeriod?period=day&month=october`
```json
GET /reports/orderedItemsByPeriod?period=day&month=january
HTTP/1.1 200 OK
Content-Type: application/json

{
    "period": "day",
    "month": "january",
    "orderedItems": [
        {
            "1": 1
        },
        {
            "2": 2
        },
        {
            "3": 1
        },
        {
            "4": 2
        },
        {
            "5": 1
        },
        {
            "6": 1
        },
        {
            "7": 2
        },
        {
            "8": 1
        },
        {
            "9": 2
        },
        {
            "10": 1
        },
        {
            "11": 1
        },
        {
            "12": 1
        },
        {
            "13": 1
        },
        {
            "14": 1
        },
        {
            "15": 1
        },
        {
            "16": 2
        },
        {
            "17": 1
        },
        {
            "18": 1
        },
        {
            "19": 1
        },
        {
            "20": 1
        },
        {
            "21": 1
        },
        {
            "22": 1
        },
        {
            "23": 1
        },
        {
            "24": 1
        },
        {
            "25": 1
        },
        {
            "26": 2
        },
        {
            "27": 1
        },
        {
            "28": 1
        },
        {
            "29": 0
        },
        {
            "30": 2
        },
        {
            "31": 1
        }
    ]
}
```

### 4. Get Inventory Leftovers  
`GET /inventory/getLeftOvers?sortBy=quantity&page=1&pageSize=10`
```json
GET /getLeftOvers?sortBy=name&page=3&pageSize=5
HTTP/1.1 200 OK
Content-Type: application/json

{
    "currentPage": 3,
    "hasNextPage": true,
    "pageSize": 5,
    "totalPages": 5,
    "data": [
        {
            "name": "Honey",
            "quantity": 1000
        },
        {
            "name": "Hazelnut Syrup",
            "quantity": 1000
        },
        {
            "name": "Ground Coffee",
            "quantity": 3000
        },
        {
            "name": "Flour",
            "quantity": 9900
        },
        {
            "name": "Espresso Shot",
            "quantity": 498
        }
    ]
}
```

### 5. Batch Order Processing  
`POST /orders/batch-process` — Handle multiple orders with inventory validation and transactions

Request body:
```json
{
    
  "orders": [
    {
      "customer_name": "John",
      "items": [
        { "menu_id": 1, "quantity": 1 },
        { "menu_id": 5, "quantity": 2 }
      ]
    },
    {
      "customer_name": "Peter",
      "items": [
        { "menu_id": 999, "quantity": 1 }
      ]
    }
  ]
}
```

Response:
```json
{
    "processed_orders": [
        {
            "order_id": 61,
            "customer_name": "John",
            "status": "accepted",
            "total": 6
        },
        {
            "order_id": 62,
            "customer_name": "Peter",
            "status": "rejected",
            "reason": "menu item does not exist"
        }
    ],
    "summary": {
        "total_orders": 2,
        "accepted": 1,
        "rejected": 1,
        "total_revenue": 6,
        "inventory_updates": [
            {
                "inventory_id": 1,
                "name": "Espresso Shot",
                "quantity_used": 2,
                "remaining": 498
            },
            {
                "inventory_id": 3,
                "name": "Flour",
                "quantity_used": 100,
                "remaining": 9900
            },
            {
                "inventory_id": 4,
                "name": "Blueberries",
                "quantity_used": 50,
                "remaining": 1950
            },
            {
                "inventory_id": 6,
                "name": "Sugar",
                "quantity_used": 10,
                "remaining": 4990
            },
            {
                "inventory_id": 15,
                "name": "Pastry Dough",
                "quantity_used": 100,
                "remaining": 4900
            },
            {
                "inventory_id": 16,
                "name": "Butter",
                "quantity_used": 20,
                "remaining": 1980
            }
        ]
    }
}
```