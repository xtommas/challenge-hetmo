# Challenge Backend

Solución para el challenge backend de Hetmo.

## Requisitos

- Gestión de eventos solamente por administradores.
- Usuarios y administradores pueden inscribirse a eventos publicados.
- Se debe poder filtrar los eventos por fecha, estado y título.
- Los usuarios pueden filtrar los eventos a los que se han inscripto en base a si ya ocurrieron o no.

## Endpoints

La API cuenta con los siguientes endpoints:

| Método | Ruta                            | Acción                           | Acceso      | Filtros                                                                                                                  |
| ------ | ------------------------------- | -------------------------------- | ----------- | ------------------------------------------------------------------------------------------------------------------------ |
| POST   | /register                       | Registro de usuario              | Público     |                                                                                                                          |
| POST   | /login                          | Login de usuario                 | Público     |                                                                                                                          |
| GET    | /api/v1/events                  | Obtener todos los eventos        | Autenticado | paginación (`page` y `limit`), `date_start` (YYYY-MM-DD), `date_end` (YYYY-MM-DD), `status` (draft o published), `title` |
| GET    | /api/v1/events/:id              | Obtener un evento específico     | Autenticado |                                                                                                                          |
| POST   | /api/v1/events                  | Crear un evento                  | Admin       |                                                                                                                          |
| DELETE | /api/v1/events/:id              | Borrar un evento                 | Admin       |                                                                                                                          |
| PATCH  | /api/v1/events/:id              | Actualizar un evento             | Admin       |                                                                                                                          |
| POST   | /api/v1/events/:id/signup       | Inscribirse a un evento          | Autenticado |                                                                                                                          |
| GET    | /api/v1/user/events             | Obtener eventos del usuario      | Autenticado | paginación (`page` y `limit`), `filter` (past o upcoming)                                                                |
| PATCH  | /api/v1/users/:username/promote | Promover usuario a administrador | Admin       |                                                                                                                          |

## Ejecución

Para ejecutar la aplicación, se debe proveer un archivo `.env` con las siguientes variables de entorno:

```
POSTGRES_DB=nombre_de_la_base_de_datos
POSTGRES_USER=usuario_de_la_base_de_datos
POSTGRES_PASSWORD=contraseña_de_la_base_de_datos
JWT_SECRET=secreto_jwt
ADMIN_USERNAME=usuario_admin
ADMIN_PASSWORD=contraseña_admin
```

Luego, se debe ejecutar el siguiente comando para iniciar la aplicación utilizando Docker:

```
docker compose up --build
```

Finalmente, se pueden realizar requests utilizando algún cliente HTTP a `localhost:8080`

## Notas y consideraciones

- No se supuso ningún orden para el listado de eventos, por lo que se devuelven sin un orden en particular.
- Los adminstradores tienen la capacidad de dar de alta un evento con el estado "published", no se fuerza a que el evento deba pasar primero por el estado "draft".
- Se asume que pueden existir múltiples adminstradores, entonces, los administradores tienen la capacidad de promover a otros usuarios a administradores utilizando su nombre de usuario.
- La aplicación crea un usuario administrador por defecto con el nombre de usuario y contraseña especificados en las variables de entorno `ADMIN_USERNAME` y `ADMIN_PASSWORD`.
