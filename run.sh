#!/bin/bash

# Script para ejecutar el servidor MCP con configuración personalizada

echo "🚀 Iniciando servidor MCP para MongoDB..."

# Cargar variables de entorno si existe el archivo .env
if [ -f .env ]; then
    echo "📋 Cargando configuración desde .env..."
    export $(grep -v '^#' .env | xargs)
else
    echo "⚠️  No se encontró archivo .env, usando configuración por defecto"
fi

# Verificar que MongoDB esté ejecutándose
if ! pgrep -x "mongod" > /dev/null; then
    echo "❌ MongoDB no está ejecutándose. Iniciando MongoDB..."
    # Intentar iniciar MongoDB (esto puede variar según el sistema)
    if command -v systemctl >/dev/null 2>&1; then
        sudo systemctl start mongod
    elif command -v brew >/dev/null 2>&1; then
        brew services start mongodb-community
    else
        echo "Por favor, inicia MongoDB manualmente"
        exit 1
    fi
    
    # Esperar un momento para que MongoDB inicie
    sleep 3
fi

echo "✅ MongoDB está ejecutándose"

# Mostrar configuración actual
echo "⚙️  Configuración del servidor:"
echo "   MongoDB URI: ${MONGODB_URI:-mongodb://localhost:27017}"
echo "   Base de datos: ${DB_NAME:-school}"
echo "   Colección: ${COLLECTION_NAME:-students}"
echo "   Puerto: ${PORT:-8080}"

# Verificar dependencias de Go
if [ ! -f "go.sum" ]; then
    echo "📦 Descargando dependencias..."
    go mod tidy
fi

# Ejecutar el servidor
echo "🎯 Iniciando servidor MCP en puerto ${PORT:-8080}..."
go run main.go