package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Estructura para representar un alumno
type Student struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name     string             `bson:"name" json:"name"`
	Subjects map[string]float64 `bson:"subjects" json:"subjects"`
}

// Estructura para el protocolo MCP
type MCPMessage struct {
	JsonRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Method  string      `json:"method,omitempty"`
	Params  interface{} `json:"params,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}

type ToolSchema struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Required   []string               `json:"required,omitempty"`
}

type Server struct {
	client     *mongo.Client
	database   *mongo.Database
	collection *mongo.Collection
}

func NewServer(mongoURI, dbName, collectionName string) (*Server, error) {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, err
	}

	// Verificar la conexión
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return nil, err
	}

	database := client.Database(dbName)
	collection := database.Collection(collectionName)

	return &Server{
		client:     client,
		database:   database,
		collection: collection,
	}, nil
}

func (s *Server) Close() error {
	return s.client.Disconnect(context.TODO())
}

// Herramientas disponibles
func (s *Server) getTools() []Tool {
	return []Tool{
		{
			Name:        "list_students",
			Description: "Lista todos los estudiantes en la base de datos",
			InputSchema: ToolSchema{
				Type:       "object",
				Properties: map[string]interface{}{},
			},
		},
		{
			Name:        "get_student_by_name",
			Description: "Busca un estudiante por su nombre",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Nombre del estudiante a buscar",
					},
				},
				Required: []string{"name"},
			},
		},
		{
			Name:        "get_student_grades",
			Description: "Obtiene las notas de un estudiante específico",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Nombre del estudiante",
					},
				},
				Required: []string{"name"},
			},
		},
		{
			Name:        "get_subject_grades",
			Description: "Obtiene todas las notas de una asignatura específica",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"subject": map[string]interface{}{
						"type":        "string",
						"description": "Nombre de la asignatura",
					},
				},
				Required: []string{"subject"},
			},
		},
		{
			Name:        "calculate_student_average",
			Description: "Calcula el promedio de notas de un estudiante",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Nombre del estudiante",
					},
				},
				Required: []string{"name"},
			},
		},
		{
			Name:        "add_student",
			Description: "Añade un nuevo estudiante a la base de datos",
			InputSchema: ToolSchema{
				Type: "object",
				Properties: map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Nombre del estudiante",
					},
					"subjects": map[string]interface{}{
						"type":        "object",
						"description": "Asignaturas y notas del estudiante (formato: {\"matematicas\": 8.5, \"historia\": 9.0})",
					},
				},
				Required: []string{"name", "subjects"},
			},
		},
	}
}

// Implementación de las herramientas
func (s *Server) listStudents() (interface{}, error) {
	cursor, err := s.collection.Find(context.TODO(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var students []Student
	if err = cursor.All(context.TODO(), &students); err != nil {
		return nil, err
	}

	return students, nil
}

func (s *Server) getStudentByName(name string) (interface{}, error) {
	var student Student
	err := s.collection.FindOne(context.TODO(), bson.M{"name": name}).Decode(&student)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("estudiante '%s' no encontrado", name)
		}
		return nil, err
	}

	return student, nil
}

func (s *Server) getStudentGrades(name string) (interface{}, error) {
	var student Student
	err := s.collection.FindOne(context.TODO(), bson.M{"name": name}).Decode(&student)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("estudiante '%s' no encontrado", name)
		}
		return nil, err
	}

	return map[string]interface{}{
		"student": name,
		"grades":  student.Subjects,
	}, nil
}

func (s *Server) getSubjectGrades(subject string) (interface{}, error) {
	cursor, err := s.collection.Find(context.TODO(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var results []map[string]interface{}
	for cursor.Next(context.TODO()) {
		var student Student
		if err := cursor.Decode(&student); err != nil {
			continue
		}

		if grade, exists := student.Subjects[subject]; exists {
			results = append(results, map[string]interface{}{
				"student": student.Name,
				"grade":   grade,
			})
		}
	}

	return map[string]interface{}{
		"subject": subject,
		"grades":  results,
	}, nil
}

func (s *Server) calculateStudentAverage(name string) (interface{}, error) {
	var student Student
	err := s.collection.FindOne(context.TODO(), bson.M{"name": name}).Decode(&student)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("estudiante '%s' no encontrado", name)
		}
		return nil, err
	}

	if len(student.Subjects) == 0 {
		return map[string]interface{}{
			"student": name,
			"average": 0,
			"message": "No hay notas registradas",
		}, nil
	}

	var total float64
	for _, grade := range student.Subjects {
		total += grade
	}
	average := total / float64(len(student.Subjects))

	return map[string]interface{}{
		"student":      name,
		"average":      average,
		"total_grades": len(student.Subjects),
	}, nil
}

func (s *Server) addStudent(name string, subjects map[string]interface{}) (interface{}, error) {
	// Convertir subjects a map[string]float64
	convertedSubjects := make(map[string]float64)
	for subject, grade := range subjects {
		switch v := grade.(type) {
		case float64:
			convertedSubjects[subject] = v
		case int:
			convertedSubjects[subject] = float64(v)
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				convertedSubjects[subject] = f
			} else {
				return nil, fmt.Errorf("nota inválida para %s: %s", subject, v)
			}
		default:
			return nil, fmt.Errorf("tipo de nota inválido para %s", subject)
		}
	}

	student := Student{
		Name:     name,
		Subjects: convertedSubjects,
	}

	result, err := s.collection.InsertOne(context.TODO(), student)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"message":    "Estudiante añadido exitosamente",
		"student_id": result.InsertedID,
		"name":       name,
		"subjects":   convertedSubjects,
	}, nil
}

func (s *Server) handleToolCall(toolName string, params map[string]interface{}) (interface{}, error) {
	switch toolName {
	case "list_students":
		return s.listStudents()
	case "get_student_by_name":
		name, ok := params["name"].(string)
		if !ok {
			return nil, fmt.Errorf("parámetro 'name' requerido")
		}
		return s.getStudentByName(name)
	case "get_student_grades":
		name, ok := params["name"].(string)
		if !ok {
			return nil, fmt.Errorf("parámetro 'name' requerido")
		}
		return s.getStudentGrades(name)
	case "get_subject_grades":
		subject, ok := params["subject"].(string)
		if !ok {
			return nil, fmt.Errorf("parámetro 'subject' requerido")
		}
		return s.getSubjectGrades(subject)
	case "calculate_student_average":
		name, ok := params["name"].(string)
		if !ok {
			return nil, fmt.Errorf("parámetro 'name' requerido")
		}
		return s.calculateStudentAverage(name)
	case "add_student":
		name, ok := params["name"].(string)
		if !ok {
			return nil, fmt.Errorf("parámetro 'name' requerido")
		}
		subjects, ok := params["subjects"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("parámetro 'subjects' requerido")
		}
		return s.addStudent(name, subjects)
	default:
		return nil, fmt.Errorf("herramienta desconocida: %s", toolName)
	}
}

// ARREGLADA: Manejo de mensajes para ambos modos (TCP y stdio)
func (s *Server) processMessage(message []byte) []byte {
	var msg MCPMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		// Si no podemos parsear el mensaje, usamos un ID por defecto
		errorResponse := MCPMessage{
			JsonRPC: "2.0",
			ID:      "error", // ID por defecto para errores de parsing
			Error: &MCPError{
				Code:    -32700,
				Message: "Parse error: " + err.Error(),
			},
		}
		responseBytes, _ := json.Marshal(errorResponse)
		return responseBytes
	}

	response := MCPMessage{
		JsonRPC: "2.0",
		ID:      msg.ID,
	}

	// Asegurar que el ID nunca sea nil
	if response.ID == nil {
		response.ID = "unknown"
	}

	switch msg.Method {
	case "initialize":
		response.Result = map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"serverInfo": map[string]interface{}{
				"name":    "mongodb-student-server",
				"version": "1.0.0",
			},
		}

	case "tools/list":
		response.Result = map[string]interface{}{
			"tools": s.getTools(),
		}

	case "tools/call":
		params, ok := msg.Params.(map[string]interface{})
		if !ok {
			response.Error = &MCPError{
				Code:    -32602,
				Message: "Parámetros inválidos",
			}
		} else {
			toolName, ok := params["name"].(string)
			if !ok {
				response.Error = &MCPError{
					Code:    -32602,
					Message: "Nombre de herramienta requerido",
				}
			} else {
				arguments, _ := params["arguments"].(map[string]interface{})
				if arguments == nil {
					arguments = make(map[string]interface{})
				}

				result, err := s.handleToolCall(toolName, arguments)
				if err != nil {
					response.Error = &MCPError{
						Code:    -32603,
						Message: err.Error(),
					}
				} else {
					response.Result = map[string]interface{}{
						"content": []map[string]interface{}{
							{
								"type": "text",
								"text": fmt.Sprintf("%+v", result),
							},
						},
					}
				}
			}
		}

	case "notifications/initialized":
		// Notificación de inicialización - no necesita respuesta
		return []byte{}

	default:
		response.Error = &MCPError{
			Code:    -32601,
			Message: "Método no encontrado: " + msg.Method,
		}
	}

	responseBytes, _ := json.Marshal(response)
	return responseBytes
}

// NUEVA FUNCIÓN: Manejo de stdin/stdout para Claude Desktop
func (s *Server) handleStdio() {
	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		// Procesar mensaje
		response := s.processMessage(line)

		// Solo enviar respuesta si no está vacía
		if len(response) > 0 {
			fmt.Printf("%s\n", response)

			// Flush stdout para asegurar que se envíe inmediatamente
			os.Stdout.Sync()
		}
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		// En modo stdio no podemos usar log porque contamina stdout
		// Solo salir silenciosamente si hay error
		os.Exit(1)
	}
}

// MODIFICADA: Usar la función compartida processMessage
func (s *Server) handleMessage(conn net.Conn, message []byte) {
	response := s.processMessage(message)
	if len(response) > 0 {
		conn.Write(response)
		conn.Write([]byte("\n"))
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	log.Printf("Nueva conexión desde %s", conn.RemoteAddr())

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) > 0 {
			s.handleMessage(conn, line)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error leyendo de la conexión: %v", err)
	}
}

func main() {
	// Configuración por defecto
	mongoURI := "mongodb://127.0.0.1:27017"
	dbName := "school"
	collectionName := "students"
	port := "8080"

	// Leer configuración desde variables de entorno si están disponibles
	if uri := os.Getenv("MONGODB_URI"); uri != "" {
		mongoURI = uri
	}
	if db := os.Getenv("DB_NAME"); db != "" {
		dbName = db
	}
	if coll := os.Getenv("COLLECTION_NAME"); coll != "" {
		collectionName = coll
	}
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	// Detectar modo de operación
	mode := getEnv("MCP_MODE", "auto")
	isStdio := mode == "stdio" || (mode == "auto" && isStdioMode())

	// Solo mostrar logs en modo TCP para no contaminar stdio
	if !isStdio {
		log.Printf("Conectando a MongoDB: %s", mongoURI)
	}

	// Crear el servidor
	server, err := NewServer(mongoURI, dbName, collectionName)
	if err != nil {
		if !isStdio {
			log.Fatalf("Error conectando a MongoDB: %v", err)
		} else {
			// En modo stdio, salir silenciosamente
			os.Exit(1)
		}
	}
	defer server.Close()

	if !isStdio {
		log.Printf("Conectado a MongoDB: %s", mongoURI)
		log.Printf("Base de datos: %s, Colección: %s", dbName, collectionName)
	}

	if isStdio {
		// Modo stdio para Claude Desktop
		server.handleStdio()
	} else {
		// Modo TCP para pruebas directas
		log.Printf("Iniciando servidor MCP en puerto %s...", port)

		// Crear el listener TCP
		listener, err := net.Listen("tcp", ":"+port)
		if err != nil {
			log.Fatalf("Error creando listener: %v", err)
		}
		defer listener.Close()

		log.Printf("Servidor MCP escuchando en puerto %s", port)

		// Aceptar conexiones
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Printf("Error aceptando conexión: %v", err)
				continue
			}

			go server.handleConnection(conn)
		}
	}
}

// NUEVAS FUNCIONES de utilidad
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func isStdioMode() bool {
	// Detectar si estamos siendo ejecutados por Claude Desktop
	// Claude Desktop no proporciona un terminal interactivo
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}

	// Si stdin es un pipe (no un terminal), probablemente estamos en modo MCP
	return (stat.Mode() & os.ModeCharDevice) == 0
}
