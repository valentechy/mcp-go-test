#!/bin/bash

# Script para inicializar la base de datos MongoDB con datos de ejemplo

echo "🔧 Configurando base de datos MongoDB..."

# Verificar si MongoDB está ejecutándose
if ! pgrep -x "mongod" > /dev/null; then
    echo "❌ MongoDB no está ejecutándose. Por favor, inicia MongoDB primero."
    exit 1
fi

echo "✅ MongoDB está ejecutándose"

# Conectar a MongoDB e insertar datos de ejemplo
mongosh --eval "
use school

// Limpiar colección existente si existe
db.students.drop()

// Insertar estudiantes de ejemplo
db.students.insertMany([
  {
    'name': 'Juan Pérez',
    'subjects': {
      'matematicas': 8.5,
      'historia': 9.0,
      'ciencias': 7.5,
      'literatura': 8.8,
      'ingles': 8.2
    }
  },
  {
    'name': 'María García',
    'subjects': {
      'matematicas': 9.2,
      'historia': 8.7,
      'ciencias': 9.5,
      'literatura': 8.9,
      'ingles': 9.1
    }
  },
  {
    'name': 'Carlos López',
    'subjects': {
      'matematicas': 7.8,
      'historia': 8.2,
      'ciencias': 8.0,
      'literatura': 7.9,
      'ingles': 8.5
    }
  },
  {
    'name': 'Ana Martínez',
    'subjects': {
      'matematicas': 9.5,
      'historia': 9.3,
      'ciencias': 9.0,
      'literatura': 9.4,
      'ingles': 9.2
    }
  },
  {
    'name': 'Luis Rodríguez',
    'subjects': {
      'matematicas': 7.2,
      'historia': 7.8,
      'ciencias': 7.5,
      'literatura': 8.1,
      'ingles': 7.9
    }
  }
])

print('📊 Datos insertados correctamente')
print('📋 Total de estudiantes:', db.students.countDocuments())
print('👥 Lista de estudiantes:')
db.students.find({}, {name: 1, _id: 0}).forEach(function(doc) {
  print('  - ' + doc.name)
})
"

echo "🎉 Base de datos configurada exitosamente!"
echo "🚀 Ahora puedes ejecutar el servidor con: go run main.go"