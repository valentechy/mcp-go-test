#!/bin/bash

# Script para inicializar la base de datos MongoDB con datos de ejemplo

echo "🔧 Configurando base de datos MongoDB..."

# Verificar si MongoDB está ejecutándose
if ! pgrep -x "mongod" > /dev/null; then
    echo "❌ MongoDB no está ejecutándose. Por favor, inicia MongoDB primero."
    exit 1
fi

echo "✅ MongoDB está ejecutándose"

# Ejecutar script de inicialización
echo "📊 Ejecutando inicializador de datos..."

# Intentar mongosh/mongo primero, luego usar Go como fallback
if command -v mongosh >/dev/null 2>&1; then
    echo "📡 Usando mongosh..."
    mongosh school sample_data.js
elif command -v mongo >/dev/null 2>&1; then
    echo "📡 Usando mongo (cliente clásico)..."
    mongo school sample_data.js
else
    echo "📡 Cliente MongoDB no encontrado, usando inicializador Go..."
    go run init_db.go
fi

echo "🎉 Base de datos configurada exitosamente!"
echo "🚀 Ahora puedes ejecutar el servidor con: go run main.go"