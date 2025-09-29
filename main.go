package main

import (
	"context"
	"encoding/json"
	"fmt"
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
	ID       primitive.ObjectID    `bson:"_id,omitempty" json:"id,omitempty"`
	Name     string               `bson:"name" json:"name"`
	Subjects map[string]float64   `bson:"subjects" json:"subjects"`
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
	Required   []string              `json:"required,omitempty"`
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

func (s *Server) handleMessage(conn net.Conn, message []byte) {
	var msg MCPMessage
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Printf("Error al parsear mensaje: %v", err)
		return
	}

	response := MCPMessage{
		JsonRPC: "2.0",
		ID:      msg.ID,
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

	default:
		response.Error = &MCPError{
			Code:    -32601,
			Message: "Método no encontrado",
		}
	}

	responseBytes, _ := json.Marshal(response)
	conn.Write(responseBytes)
	conn.Write([]byte("\n"))
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	log.Printf("Nueva conexión desde %s", conn.RemoteAddr())

	buffer := make([]byte, 4096)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			log.Printf("Error leyendo de la conexión: %v", err)
			break
		}

		message := buffer[:n]
		s.handleMessage(conn, message)
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

	// Crear el servidor
	server, err := NewServer(mongoURI, dbName, collectionName)
	if err != nil {
		log.Fatalf("Error conectando a MongoDB: %v", err)
	}
	defer server.Close()

	log.Printf("Conectado a MongoDB: %s", mongoURI)
	log.Printf("Base de datos: %s, Colección: %s", dbName, collectionName)

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