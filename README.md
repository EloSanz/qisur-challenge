# Qisur Service

API RESTful y WebSocket para gestión de productos y categorías en tiempo real, desarrollado como parte del Qisur Challenge.

## Características

- **API REST**: Operaciones CRUD para Productos y Categorías con filtros, paginación e historial de precios/stock.
- **WebSockets**: Eventos en tiempo real para mutaciones de entidades.
- **Autenticación**: Endpoints protegidos con JWT y roles (`admin`, `client`).
- **Base de Datos**: PostgreSQL gestionado con GORM (relaciones Muchos a Muchos y transacciones para reducción de stock y logueo histórico).
- **Arquitectura**: Domain-Driven Design simplificado y Clean Architecture, acatando `go_architecture_guidelines.md`.

## Estructura de Directorios

```
qisur-service/
├── cmd/api/main.go          # Entrypoint de la app
├── internal/
│   ├── api/                 # Enrutamiento Gin
│   ├── auth/                # JWT y Middlewares
│   ├── bootstrap/           # Inyección de Dependencias
│   ├── category/            # Dominio de Categorías
│   ├── database/            # Conexión GORM
│   ├── product/             # Dominio de Productos
│   ├── search/              # Lógica de Búsqueda
│   └── websocket/           # Hub WS y Clientes
└── pkg/web/                 # Respuestas estándar JSON
```

## Setup y Configuración

1. Asegúrate de tener Docker instalado.
2. Abre una terminal y navega directamente al directorio del microservicio:
   ```bash
   cd qisur-service
   ```
3. Ejecuta el entorno aislado (con Grafana, Loki y BD propias):
   ```bash
   docker-compose up -d --build
   ```
   *Esto inicializará la base de datos `qisur_db` e instalará las tablas, además de levantar el microservicio.*

4. El servicio se levantará en el puerto `8086`. Grafana estará disponible en el puerto `3001` para revisar logs.

## Uso de la API REST

### Autenticación
Genera un token JWT (simulado para el challenge):
```bash
curl -X POST http://localhost:8086/api/auth/login \
     -H "Content-Type: application/json" \
     -d '{"username": "admin", "password": "123"}'
```
Recibirás un token. Úsalo como header `Authorization: Bearer <TOKEN>` en las siguientes peticiones protegidas.

### Endpoints (Protegidos)
*Nota: `POST`, `PUT` y `DELETE` requieren rol `admin`.*

- `GET /api/products?page=1&pageSize=10`
- `GET /api/products/:id`
- `POST /api/products` (Payload: `name`, `price`, `stock`, `category_ids[]`)
- `PUT /api/products/:id`
- `DELETE /api/products/:id`
- `GET /api/products/:id/history`

- `GET /api/categories`
- `POST /api/categories`
- `PUT /api/categories/:id`
- `DELETE /api/categories/:id`

- `GET /api/search?type=product&q=monitor`

## WebSockets (Tiempo Real)

Puedes escuchar eventos en tiempo real conectándote al endpoint público:
```bash
wscat -c ws://localhost:8086/ws
```

**Ejemplo de Evento Emitido:**
```json
{
  "event": "PRODUCT_CREATED",
  "data": {
    "id": "uuid-v4...",
    "name": "Nuevo Producto",
    "price": 100.5,
    "stock": 50,
    "categories": []
  }
}
```
