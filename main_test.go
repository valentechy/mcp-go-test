package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"testing"
)

// Test helper para conectar al servidor
func connectToServer(t *testing.T, port string) net.Conn {
	conn, err := net.Dial("tcp", "localhost:"+port)
	if err != nil {
		t.Fatalf("Error conectando al servidor: %v", err)
	}
	return conn
}

// Test helper para enviar mensaje MCP
func sendMCPMessage(conn net.Conn, message MCPMessage) (MCPMessage, error) {
	// Enviar mensaje
	data, err := json.Marshal(message)
	if err != nil {
		return MCPMessage{}, err
	}

	_, err = conn.Write(data)
	if err != nil {
		return MCPMessage{}, err
	}

	// Leer respuesta
	buffer := make([]byte, 4096)
	n, err := conn.Read(buffer)
	if err != nil {
		return MCPMessage{}, err
	}

	var response MCPMessage
	err = json.Unmarshal(bytes.TrimSpace(buffer[:n]), &response)
	return response, err
}

func TestServerInitialize(t *testing.T) {
	// Esta prueba requiere que el servidor esté ejecutándose
	t.Skip("Requiere servidor ejecutándose - ejecutar manualmente")

	conn := connectToServer(t, "8080")
	defer conn.Close()

	initMessage := MCPMessage{
		JsonRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params:  map[string]interface{}{},
	}

	response, err := sendMCPMessage(conn, initMessage)
	if err != nil {
		t.Fatalf("Error enviando mensaje de inicialización: %v", err)
	}

	if response.Error != nil {
		t.Fatalf("Error en respuesta: %v", response.Error)
	}

	result, ok := response.Result.(map[string]interface{})
	if !ok {
		t.Fatal("Respuesta no tiene formato esperado")
	}

	if result["protocolVersion"] != "2024-11-05" {
		t.Errorf("Versión de protocolo incorrecta: %v", result["protocolVersion"])
	}
}

func TestListTools(t *testing.T) {
	// Esta prueba requiere que el servidor esté ejecutándose
	t.Skip("Requiere servidor ejecutándose - ejecutar manualmente")

	conn := connectToServer(t, "8080")
	defer conn.Close()

	message := MCPMessage{
		JsonRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	}

	response, err := sendMCPMessage(conn, message)
	if err != nil {
		t.Fatalf("Error enviando mensaje: %v", err)
	}

	if response.Error != nil {
		t.Fatalf("Error en respuesta: %v", response.Error)
	}

	result, ok := response.Result.(map[string]interface{})
	if !ok {
		t.Fatal("Respuesta no tiene formato esperado")
	}

	tools, ok := result["tools"].([]interface{})
	if !ok {
		t.Fatal("Tools no tiene formato esperado")
	}

	expectedTools := []string{
		"list_students",
		"get_student_by_name",
		"get_student_grades",
		"get_subject_grades",
		"calculate_student_average",
		"add_student",
	}

	if len(tools) != len(expectedTools) {
		t.Errorf("Número de herramientas incorrecto. Esperado: %d, Obtenido: %d",
			len(expectedTools), len(tools))
	}
}

// Test unitario para la estructura Student
func TestStudentStruct(t *testing.T) {
	student := Student{
		Name: "Test Student",
		Subjects: map[string]float64{
			"math":    8.5,
			"science": 9.0,
		},
	}

	if student.Name != "Test Student" {
		t.Errorf("Nombre incorrecto: %s", student.Name)
	}

	if len(student.Subjects) != 2 {
		t.Errorf("Número de asignaturas incorrecto: %d", len(student.Subjects))
	}

	if student.Subjects["math"] != 8.5 {
		t.Errorf("Nota de matemáticas incorrecta: %f", student.Subjects["math"])
	}
}

// Test para validar la serialización JSON
func TestMCPMessageSerialization(t *testing.T) {
	msg := MCPMessage{
		JsonRPC: "2.0",
		ID:      1,
		Method:  "test",
		Params:  map[string]interface{}{"key": "value"},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Error serializando mensaje: %v", err)
	}

	var deserialized MCPMessage
	err = json.Unmarshal(data, &deserialized)
	if err != nil {
		t.Fatalf("Error deserializando mensaje: %v", err)
	}

	if deserialized.JsonRPC != msg.JsonRPC {
		t.Errorf("JsonRPC incorrecto: %s", deserialized.JsonRPC)
	}

	if deserialized.Method != msg.Method {
		t.Errorf("Method incorrecto: %s", deserialized.Method)
	}
}

// Benchmark para la serialización de mensajes
func BenchmarkMessageSerialization(b *testing.B) {
	msg := MCPMessage{
		JsonRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      "list_students",
			"arguments": map[string]interface{}{},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(msg)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Test de integración (requiere MongoDB y servidor ejecutándose)
func TestIntegrationListStudents(t *testing.T) {
	t.Skip("Test de integración - ejecutar manualmente con servidor corriendo")

	conn := connectToServer(t, "8080")
	defer conn.Close()

	// Primero inicializar
	initMsg := MCPMessage{
		JsonRPC: "2.0",
		ID:      1,
		Method:  "initialize",
	}

	_, err := sendMCPMessage(conn, initMsg)
	if err != nil {
		t.Fatalf("Error inicializando: %v", err)
	}

	// Luego llamar a list_students
	listMsg := MCPMessage{
		JsonRPC: "2.0",
		ID:      2,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      "list_students",
			"arguments": map[string]interface{}{},
		},
	}

	response, err := sendMCPMessage(conn, listMsg)
	if err != nil {
		t.Fatalf("Error llamando list_students: %v", err)
	}

	if response.Error != nil {
		t.Fatalf("Error en respuesta: %v", response.Error)
	}

	fmt.Printf("Respuesta de list_students: %+v\n", response.Result)
}
