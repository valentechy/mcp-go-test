// Datos de ejemplo compartidos para inicialización
const sampleStudents = [
  {
    "name": "Juan Pérez",
    "subjects": {
      "matematicas": 8.5,
      "historia": 9.0,
      "ciencias": 7.5,
      "literatura": 8.8,
      "ingles": 8.2
    }
  },
  {
    "name": "María García",
    "subjects": {
      "matematicas": 9.2,
      "historia": 8.7,
      "ciencias": 9.5,
      "literatura": 8.9,
      "ingles": 9.1
    }
  },
  {
    "name": "Carlos López",
    "subjects": {
      "matematicas": 7.8,
      "historia": 8.2,
      "ciencias": 8.0,
      "literatura": 7.9,
      "ingles": 8.5
    }
  },
  {
    "name": "Ana Martínez",
    "subjects": {
      "matematicas": 9.5,
      "historia": 9.3,
      "ciencias": 9.0,
      "literatura": 9.4,
      "ingles": 9.2
    }
  },
  {
    "name": "Luis Rodríguez",
    "subjects": {
      "matematicas": 7.2,
      "historia": 7.8,
      "ciencias": 7.5,
      "literatura": 8.1,
      "ingles": 7.9
    }
  }
];

// Función para insertar estudiantes
function insertSampleStudents() {
  const result = db.students.insertMany(sampleStudents);
  print(`✅ Insertados ${result.insertedIds.length} estudiantes de ejemplo`);
  
  // Mostrar estadísticas
  print('📊 Estudiantes en la base de datos:');
  db.students.find({}, {name: 1, _id: 0}).forEach(function(doc) {
    print(`  - ${doc.name}`);
  });
}

// Para uso en MongoDB local (mongosh)
if (typeof module === 'undefined') {
  // Limpiar colección existente
  db.students.drop();
  print('🧹 Colección students limpiada');
  
  // Insertar datos
  insertSampleStudents();
}

// Para exportar en Node.js/Docker si es necesario
if (typeof module !== 'undefined') {
  module.exports = { sampleStudents, insertSampleStudents };
}