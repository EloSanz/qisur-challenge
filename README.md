# Qisur Challenge

Solución completa para el challenge de Qisur, que incluye una API RESTful, WebSockets, procesamiento asíncrono con RabbitMQ y un Frontend en React para monitorear los eventos en tiempo real.

Toda la infraestructura está dockerizada y actualmente desplegada en producción a través de un VPS con Nginx como proxy inverso.

## 🚀 Entorno de Producción (Live Demo)

Todo el proyecto está expuesto públicamente. Puedes probarlo usando los siguientes accesos:

- **Frontend (Tracker UI):** [https://icards.fun/qisur/](https://icards.fun/qisur/)
- **Documentación Swagger:** [https://icards.fun/qisur/swagger/index.html](https://icards.fun/qisur/swagger/index.html)
- **API Base URL:** `https://icards.fun/qisur/api`
- **WebSockets:** `wss://icards.fun/qisur/ws/`

---

## 🛠️ Cómo Probar la Aplicación

Para ver la aplicación en acción y visualizar cómo el Tracker atrapa los eventos en tiempo real, sigue estos pasos:

### 1. Abre el Frontend (Tracker UI)
Abre [https://icards.fun/qisur/](https://icards.fun/qisur/) en una pestaña de tu navegador. Esta interfaz React está conectada directamente al Backend para listar las trazas de auditoría (Audit Traces). Déjala abierta.

### 2. Autentícate en Swagger
Abre la [Documentación Swagger](https://icards.fun/qisur/swagger/index.html) en otra pestaña. 

Para poder realizar operaciones de escritura (Crear productos, categorías, etc.), necesitas un **Token de Administrador**.
1. En Swagger, busca el endpoint `POST /api/auth/login`.
2. Haz clic en **Try it out**.
3. El JSON ya vendrá pre-llenado con las credenciales maestras:
   ```json
   {
     "username": "admin",
     "password": "admin"
   }
   ```
4. Haz clic en **Execute**. (Nota: El sistema está mockeado; al detectar la palabra `admin` te devuelve un JWT con el rol de Administrador).
5. En la respuesta, copia el string del token generado.
6. Sube al inicio de la página y haz clic en el botón verde **Authorize**.
7. En la caja de texto, pega tu token **anteponiendo la palabra Bearer**:
   ```text
   Bearer eyJhbGciOiJIUzI1NiIs...
   ```
8. Haz clic en Authorize y luego en Close.

### 3. Dispara Eventos (Impulsos)
Ahora que estás autenticado, puedes usar cualquier endpoint protegido, por ejemplo:
- `POST /api/categories` (Crear una categoría)
- `POST /api/products` (Crear un producto)
- `PUT /api/products/{id}` (Actualizar un producto)

Llena los datos en el Swagger y haz clic en **Execute**.

### 4. Observa la Magia en el Tracker
Vuelve a la pestaña del **Frontend (Tracker UI)**. Verás que en cuanto Swagger te confirma el 200/201, el Tracker UI recibe la notificación en **tiempo real**, procesada asíncronamente a través de la arquitectura de la aplicación (Golang API -> RabbitMQ -> Golang Tracker -> Postgres -> React UI).

---

## 🏗️ Arquitectura del Sistema

El ecosistema está compuesto por múltiples contenedores trabajando en conjunto:

- **qisur-service (API Gateway / Core):** API REST principal construida en **Golang (Gin)**. Gestiona la lógica de productos/categorías, guarda en PostgreSQL, publica eventos en RabbitMQ y sirve WebSockets.
- **qisur-tracker (Audit Service):** Microservicio en **Golang** sin exposición HTTP directa. Consume la cola de RabbitMQ y persiste los eventos de auditoría en la base de datos de tracking.
- **qisur-gateway (Nginx Interno):** Orquesta el tráfico de la red Docker.
- **RabbitMQ:** Message Broker para comunicación asíncrona entre microservicios.
- **PostgreSQL:** Base de datos relacional compartida.
- **Tracker UI:** Aplicación estática **React** consumiendo la API y mostrando las trazas.

## 💻 Desarrollo Local

Para correr todo el stack de forma local:

1. Clona el repositorio.
2. Asegúrate de tener Docker y Docker Compose instalados.
3. Ejecuta:
   ```bash
   docker-compose up -d --build
   ```
4. El Swagger local estará disponible en `http://localhost:8080/qisur/swagger/index.html`.
