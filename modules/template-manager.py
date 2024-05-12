from flask import Flask, request, jsonify
from pymongo import MongoClient
from bson.objectid import ObjectId

app = Flask(__name__)

# Configuración de la conexión a MongoDB
client = MongoClient('localhost', 27017)
db = client.cloud
collection = db.templates


# Función para serializar documentos MongoDB / necesario para ser enviados como respuesta del endpoint / working
def serialize_document(doc):
    for key, value in doc.items():
        if isinstance(value, ObjectId):
            doc[key] = str(value)  # Convertir ObjectId a cadena de texto
    return doc

# Endpoint para listar todas las plantillas / working
@app.route('/templates', methods=['GET'])
def listar_plantillas():
    templates = [serialize_document(template) for template in collection.find()]
    return jsonify(templates)

# Endpoint para crear una nueva plantilla / working
@app.route('/templates', methods=['POST'])
def crear_plantilla():
    nueva_plantilla = request.json
    result = collection.insert_one(nueva_plantilla)
    return jsonify({'id': str(result.inserted_id), 'mensaje': 'Plantilla creada correctamente'})

# Endpoint para editar una plantilla por ID / verify later
@app.route('/templates/<string:template_id>', methods=['PUT'])
def editar_plantilla(template_id):
    plantilla_actualizada = request.json
    result = collection.update_one({'_id': ObjectId(template_id)}, {'$set': plantilla_actualizada})
    if result.modified_count == 1:
        return jsonify({'mensaje': 'Plantilla actualizada correctamente'})
    else:
        return jsonify({'error': 'La plantilla no se pudo actualizar'})

# Endpoint para eliminar una plantilla por ID / working
@app.route('/templates/<string:template_id>', methods=['DELETE'])
def eliminar_plantilla(template_id):
    result = collection.delete_one({'_id': ObjectId(template_id)})
    if result.deleted_count == 1:
        return jsonify({'mensaje': 'Plantilla eliminada correctamente'})
    else:
        return jsonify({'error': 'La plantilla no se pudo eliminar'})

# Endpoint para buscar una plantilla por ID / working
@app.route('/templates/<string:template_id>', methods=['GET'])
def buscar_plantilla(template_id):
    plantilla = collection.find_one({'_id': ObjectId(template_id)})
    if plantilla:
        plantilla_serialized = serialize_document(plantilla)
        return jsonify(plantilla_serialized)
    else:
        return jsonify({'error': 'Plantilla no encontrada'})

if __name__ == '__main__':
    app.run(debug=True)