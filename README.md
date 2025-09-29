# Servidor MCP para MongoDB - Gestión de Estudiantes

Este proyecto implementa un servidor MCP (Model Context Protocol) en Go para consultar una base de datos MongoDB que contiene información de estudiantes y sus notas por asignaturas.

## Características

- **Conexión a MongoDB local**: Se conecta a una instancia de MongoDB en localhost
- **Gestión de estudiantes**: CRUD básico para estudiantes y sus notas
- **Consultas avanzadas**: Búsquedas por nombre, asignatura, cálculo de promedios
- **Protocolo MCP**: Implementa el protocolo estándar MCP para integración con herramientas de IA

## Estructura de Datos

Cada estudiante se almacena con la siguiente estructura:

```json
{
  "_id": "ObjectId",
  "name": "Juan Pérez",
  "subjects": {
    "matematicas": 8.5,
    "historia": 9.0,
    "ciencias": 7.5,
    "literatura": 8.8
  }
}
```

## Herramientas Disponibles

El servidor MCP proporciona las siguientes herramientas:

1. **`list_students`**: Lista todos los estudiantes en la base de datos
2. **`get_student_by_name`**: Busca un estudiante por su nombre
3. **`get_student_grades`**: Obtiene las notas de un estudiante específico
4. **`get_subject_grades`**: Obtiene todas las notas de una asignatura
5. **`calculate_student_average`**: Calcula el promedio de notas de un estudiante
6. **`add_student`**: Añade un nuevo estudiante con sus notas

## Configuración

### Variables de Entorno

Puedes configurar el servidor usando las siguientes variables de entorno:

- `MONGODB_URI`: URI de conexión a MongoDB (por defecto: `mongodb://localhost:27017`)
- `DB_NAME`: Nombre de la base de datos (por defecto: `school`)
- `COLLECTION_NAME`: Nombre de la colección (por defecto: `students`)
- `PORT`: Puerto del servidor MCP (por defecto: `8080`)

### Ejemplo de configuración:

```bash
export MONGODB_URI="mongodb://localhost:27017"
export DB_NAME="school"
export COLLECTION_NAME="students"
export PORT="8080"
```

## Instalación y Uso

### Prerrequisitos

1. **Go 1.21+** instalado
2. **MongoDB** ejecutándose localmente en el puerto 27017
3. Una base de datos llamada `school` con una colección `students`

### Instalación

1. Clona el repositorio:
```bash
git clone <tu-repositorio>
cd mcp-go-test
```

2. Descarga las dependencias:
```bash
go mod tidy
```

3. Ejecuta el servidor:
```bash
go run main.go
```

### Configuración de MongoDB

Para preparar tu base de datos MongoDB, puedes usar estos comandos en la consola de MongoDB:

```javascript
// Conectar a la base de datos
use school

// Insertar algunos estudiantes de ejemplo
db.students.insertMany([
  {
    "name": "Juan Pérez",
    "subjects": {
      "matematicas": 8.5,
      "historia": 9.0,
      "ciencias": 7.5,
      "literatura": 8.8
    }
  },
  {
    "name": "María García",
    "subjects": {
      "matematicas": 9.2,
      "historia": 8.7,
      "ciencias": 9.5,
      "literatura": 8.9
    }
  },
  {
    "name": "Carlos López",
    "subjects": {
      "matematicas": 7.8,
      "historia": 8.2,
      "ciencias": 8.0,
      "literatura": 7.9
    }
  }
])

// Verificar que se insertaron correctamente
db.students.find().pretty()
```

## Uso del Servidor MCP

Una vez que el servidor esté ejecutándose, puedes conectarte a él usando cualquier cliente MCP en el puerto configurado (por defecto 8080).

### Ejemplo de uso con herramientas:

1. **Listar estudiantes**:
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "list_students"
  }
}
```

2. **Buscar un estudiante**:
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/call",
  "params": {
    "name": "get_student_by_name",
    "arguments": {
      "name": "Juan Pérez"
    }
  }
}
```

3. **Calcular promedio**:
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "calculate_student_average",
    "arguments": {
      "name": "María García"
    }
  }
}
```

## Desarrollo

### Estructura del proyecto

```
mcp-go-test/
├── main.go          # Archivo principal del servidor
├── go.mod           # Dependencias de Go
├── go.sum           # Checksums de dependencias
└── README.md        # Este archivo
```

### Próximas características

- [ ] Autenticación y autorización
- [ ] Más operaciones CRUD (actualizar, eliminar estudiantes)
- [ ] Filtros avanzados por rango de notas
- [ ] Estadísticas por clase/grupo
- [ ] Exportación de datos
- [ ] Logging más detallado
- [ ] Tests unitarios

## Contribuir

1. Fork el proyecto
2. Crea una rama para tu feature (`git checkout -b feature/AmazingFeature`)
3. Commit tus cambios (`git commit -m 'Add some AmazingFeature'`)
4. Push a la rama (`git push origin feature/AmazingFeature`)
5. Abre un Pull Request

## Licencia

Este proyecto está bajo la licencia MIT. Ver el archivo `LICENSE` para más detalles.

## Soporte

Si tienes problemas o preguntas:

1. Revisa que MongoDB esté ejecutándose correctamente
2. Verifica que las variables de entorno estén configuradas
3. Consulta los logs del servidor para mensajes de error
4. Abre un issue en el repositorio con detalles del problema