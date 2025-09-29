#!/bin/bash

# Script para probar el servidor MCP con ejemplos de consultas

SERVER_HOST="127.0.0.1"
SERVER_PORT="8080"

echo "🧪 Script de pruebas para el servidor MCP MongoDB"
echo "📡 Servidor: ${SERVER_HOST}:${SERVER_PORT}"
echo ""

# Función para enviar mensajes MCP usando netcat
send_mcp_message() {
    local message="$1"
    echo "📤 Enviando: $message"
    echo "$message" | nc $SERVER_HOST $SERVER_PORT
    echo ""
}

# Verificar que netcat esté disponible
if ! command -v nc &> /dev/null; then
    echo "❌ netcat (nc) no está instalado. Instalándolo..."
    if command -v apt-get &> /dev/null; then
        sudo apt-get update && sudo apt-get install -y netcat
    elif command -v yum &> /dev/null; then
        sudo yum install -y nc
    elif command -v brew &> /dev/null; then
        brew install netcat
    else
        echo "Por favor instala netcat manualmente"
        exit 1
    fi
fi

# Verificar que el servidor esté corriendo
if ! nc -z $SERVER_HOST $SERVER_PORT; then
    echo "❌ El servidor no está corriendo en ${SERVER_HOST}:${SERVER_PORT}"
    echo "Por favor ejecuta: ./run.sh"
    exit 1
fi

echo "✅ Servidor encontrado, ejecutando pruebas..."
echo ""

# 1. Inicializar el servidor
echo "1️⃣ Inicializando servidor..."
send_mcp_message '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {}
}'

# 2. Listar herramientas disponibles
echo "2️⃣ Listando herramientas disponibles..."
send_mcp_message '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/list"
}'

# 3. Listar todos los estudiantes
echo "3️⃣ Listando todos los estudiantes..."
send_mcp_message '{
    "jsonrpc": "2.0",
    "id": 3,
    "method": "tools/call",
    "params": {
        "name": "list_students"
    }
}'

# 4. Buscar un estudiante específico
echo "4️⃣ Buscando estudiante María García..."
send_mcp_message '{
    "jsonrpc": "2.0",
    "id": 4,
    "method": "tools/call",
    "params": {
        "name": "get_student_by_name",
        "arguments": {
            "name": "María García"
        }
    }
}'

# 5. Obtener notas de un estudiante
echo "5️⃣ Obteniendo notas de Juan Pérez..."
send_mcp_message '{
    "jsonrpc": "2.0",
    "id": 5,
    "method": "tools/call",
    "params": {
        "name": "get_student_grades",
        "arguments": {
            "name": "Juan Pérez"
        }
    }
}'

# 6. Obtener notas de una asignatura
echo "6️⃣ Obteniendo notas de matemáticas..."
send_mcp_message '{
    "jsonrpc": "2.0",
    "id": 6,
    "method": "tools/call",
    "params": {
        "name": "get_subject_grades",
        "arguments": {
            "subject": "matematicas"
        }
    }
}'

# 7. Calcular promedio de un estudiante
echo "7️⃣ Calculando promedio de Ana Martínez..."
send_mcp_message '{
    "jsonrpc": "2.0",
    "id": 7,
    "method": "tools/call",
    "params": {
        "name": "calculate_student_average",
        "arguments": {
            "name": "Ana Martínez"
        }
    }
}'

# 8. Añadir un nuevo estudiante
echo "8️⃣ Añadiendo nuevo estudiante..."
send_mcp_message '{
    "jsonrpc": "2.0",
    "id": 8,
    "method": "tools/call",
    "params": {
        "name": "add_student",
        "arguments": {
            "name": "Pedro Sánchez",
            "subjects": {
                "matematicas": 8.8,
                "historia": 9.2,
                "ciencias": 8.5,
                "literatura": 8.7,
                "ingles": 9.0
            }
        }
    }
}'

echo "🎉 Pruebas completadas!"
echo ""
echo "💡 Puedes ejecutar consultas personalizadas usando:"
echo "   echo '{\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"tools/call\",\"params\":{\"name\":\"HERRAMIENTA\",\"arguments\":{...}}}' | nc $SERVER_HOST $SERVER_PORT"