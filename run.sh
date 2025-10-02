#!/bin/bash
# filepath: /home/surver/001_IT/001_Personal_Projects/030_MCP-Go-Test/run.sh

# Script para ejecutar el servidor MCP con configuración personalizada
cd "$(dirname "$0")"

# Detectar la ruta de Go
GO_PATH=""
if command -v go >/dev/null 2>&1; then
    GO_PATH="go"
elif [ -f "/usr/local/go/bin/go" ]; then
    GO_PATH="/usr/local/go/bin/go"
elif [ -f "/usr/bin/go" ]; then
    GO_PATH="/usr/bin/go"
elif [ -f "$HOME/go/bin/go" ]; then
    GO_PATH="$HOME/go/bin/go"
else
    # Buscar Go en rutas comunes
    for path in /usr/local/go/bin /opt/go/bin /snap/bin ~/.local/bin; do
        if [ -f "$path/go" ]; then
            GO_PATH="$path/go"
            break
        fi
    done
fi

# Si aún no encontramos Go, intentar con PATH extendido
if [ -z "$GO_PATH" ]; then
    export PATH="/usr/local/go/bin:/opt/go/bin:/snap/bin:$HOME/.local/bin:$PATH"
    if command -v go >/dev/null 2>&1; then
        GO_PATH="go"
    fi
fi

# Detectar si estamos siendo ejecutados por Claude Desktop
if [ -t 0 ] && [ -z "$MCP_MODE" ]; then
    # stdin es un terminal y no hay MCP_MODE forzado (modo TCP)
    IS_STDIO_MODE=false
    export MCP_MODE="tcp"
elif [ "$MCP_MODE" = "tcp" ]; then
    # Forzado a modo TCP
    IS_STDIO_MODE=false
else
    # stdin es un pipe o MCP_MODE forzado a stdio (Claude Desktop)
    IS_STDIO_MODE=true
    export MCP_MODE="stdio"
fi

# Verificar que Go esté disponible
if [ -z "$GO_PATH" ]; then
    if [ "$IS_STDIO_MODE" = false ]; then
        echo "❌ Error: No se pudo encontrar Go. Por favor instala Go o añádelo al PATH" >&2
        echo "   Rutas buscadas:" >&2
        echo "   - /usr/local/go/bin/go" >&2
        echo "   - /usr/bin/go" >&2
        echo "   - /opt/go/bin/go" >&2
        echo "   - ~/.local/bin/go" >&2
        exit 1
    else
        # En modo stdio, salir silenciosamente
        exit 1
    fi
fi

# Solo mostrar mensajes en modo TCP (para no interferir con Claude Desktop)
if [ "$IS_STDIO_MODE" = false ]; then
    echo "🚀 Iniciando servidor MCP para MongoDB (modo: $MCP_MODE)..." >&2
    echo "📍 Usando Go en: $GO_PATH" >&2

    # Cargar variables de entorno si existe el archivo .env
    if [ -f .env ]; then
        echo "📋 Cargando configuración desde .env..." >&2
        export $(grep -v '^#' .env | xargs)
    else
        echo "⚠️  No se encontró archivo .env, usando configuración por defecto" >&2
    fi

    # Verificar que MongoDB esté ejecutándose
    if ! pgrep -x "mongod" > /dev/null; then
        echo "❌ MongoDB no está ejecutándose. Iniciando MongoDB..." >&2
        # Intentar iniciar MongoDB (esto puede variar según el sistema)
        if command -v systemctl >/dev/null 2>&1; then
            sudo systemctl start mongod 2>/dev/null
        elif command -v brew >/dev/null 2>&1; then
            brew services start mongodb-community 2>/dev/null
        else
            echo "Por favor, inicia MongoDB manualmente" >&2
            # No hacer exit en modo TCP para debugging
        fi
        
        # Esperar un momento para que MongoDB inicie
        sleep 2
    fi

    echo "✅ MongoDB está ejecutándose" >&2

    # Mostrar configuración actual
    echo "⚙️  Configuración del servidor:" >&2
    echo "   Modo: $MCP_MODE" >&2
    echo "   MongoDB URI: ${MONGODB_URI:-mongodb://127.0.0.1:27017}" >&2
    echo "   Base de datos: ${DB_NAME:-school}" >&2
    echo "   Colección: ${COLLECTION_NAME:-students}" >&2
    if [ "$MCP_MODE" = "tcp" ]; then
        echo "   Puerto: ${PORT:-8080}" >&2
    fi

    # Verificar dependencias de Go
    if [ ! -f "go.sum" ]; then
        echo "📦 Descargando dependencias..." >&2
        "$GO_PATH" mod tidy
    fi

    if [ "$MCP_MODE" = "tcp" ]; then
        echo "🎯 Iniciando servidor MCP en puerto ${PORT:-8080}..." >&2
    else
        echo "🎯 Iniciando servidor MCP en modo stdio..." >&2
    fi
else
    # Modo silencioso para Claude Desktop - solo cargar configuración
    if [ -f .env ]; then
        export $(grep -v '^#' .env | xargs 2>/dev/null)
    fi
    
    # Verificar dependencias silenciosamente
    if [ ! -f "go.sum" ]; then
        "$GO_PATH" mod tidy 2>/dev/null
    fi
fi

# Ejecutar el servidor
exec "$GO_PATH" run main.go