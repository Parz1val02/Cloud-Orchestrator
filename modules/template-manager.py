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
    doc['template_id'] = str(doc.pop('_id'))
    for key, value in doc.items():
        if isinstance(value, ObjectId):
            #  Convertir ObjectId a cadena de texto
            doc[key] = str(value)  
    return doc

# Endpoint para listar todas las plantillas / working
@app.route('/templates', methods=['GET'])
def listar_plantillas():

    query = {}
    fields = {'name': 1, 'description': 1, 'created_at': 1, '_id': 1}
    templates = [serialize_document(template) for template in collection.find(query,fields)]
    if templates:
        return jsonify({'result': 'success', 'templates': templates})
    else:
        return jsonify({'result': 'success', 'msg': 'No existen templates vinculados al usuario'})

    """
    role = request.args.get('role')
    if not role:
        response = jsonify({'result': 'error', 'msg': 'role es requerido'})
        error_code = 400
        return response, error_code
    else:
        fields = {'name': 1, 'description': 1, 'created_at': 1, '_id': 1}
        if role=='user':
            user_id = request.args.get('user_id')
            query = {'user_id': user_id}
            if not user_id:
                response = jsonify({'result': 'error', 'msg': 'user_id es requerido'})
                error_code = 400
                return response, error_code
            templates = [serialize_document(template) for template in collection.find(query,fields)]
            if templates:
                return jsonify({'result': 'success', 'templates': templates})
            else:
                return jsonify({'result': 'success', 'msg': 'No existen templates vinculados al usuario'})
        
        elif (role=='administrator'):
            query = {}
            templates = [serialize_document(template) for template in collection.find(query,fields)]
            if templates:
                return jsonify({'result': 'success', 'templates': templates})
            else:
                return jsonify({'result': 'success', 'msg': 'No existen templates vinculados al usuario'})
        else:
            response = jsonify({'result': 'error', 'msg': 'invalid role'})
            error_code = 400
            return response, error_code
    """

# Endpoint para buscar una plantilla por ID / working
@app.route('/templates/<string:template_id>', methods=['GET'])
def buscar_plantilla(template_id):
    try:
        template = collection.find_one({'_id': ObjectId(template_id)})
        if template:
            template_serialized = serialize_document(template)
            response = jsonify({'result': 'success', 'template': template_serialized})
            return response
        else:
            response =  jsonify({'result': 'error','msg': f'template with template_id {template_id} not found'})
            error_code = 404
            return response, error_code
    except:
            response =  jsonify({'result': 'error','msg': f'invalid template_id {template_id}'})
            error_code = 400
            return response, error_code

    

# Endpoint para crear una nueva plantilla / working
@app.route('/templates', methods=['POST'])
def crear_plantilla():
    new_template = request.json
    result = collection.insert_one(new_template)
    return jsonify({'template_id': str(result.inserted_id), 'msg': 'Plantilla creada correctamente','result':'success'})

# Endpoint para editar una plantilla por ID / working
@app.route('/templates/<string:template_id>', methods=['PUT'])
def editar_plantilla(template_id):
    plantilla_actualizada = request.json
    
    try:
        result = collection.update_one({'_id': ObjectId(template_id)}, {'$set': plantilla_actualizada})
        if result.modified_count == 1:
            return jsonify({'result': 'success' ,'msg': f'Plantilla con template_id {template_id} actualizada correctamente'})
        else:
            response = jsonify({'result': 'error', 'msg': f'Plantilla con template_id {template_id} no se pudo actualizar'})
            error_code = 400
            return response
    except:
        response = jsonify({'result': 'error', 'msg': f'template_id {template_id} inválido'})
        error_code = 400
        return response, error_code

    
# Endpoint para eliminar una plantilla por ID / working
@app.route('/templates/<string:template_id>', methods=['DELETE'])
def eliminar_plantilla(template_id):
    try:
        result = collection.delete_one({'_id': ObjectId(template_id)})
        if result.deleted_count == 1:
            return jsonify({'result':'success', 'msg': f'Plantilla con template_id {template_id} eliminada correctamente'})
        else:
            response = jsonify({'result': 'error', 'msg': f'Plantilla con template_id {template_id} no se pudo borrar'})
            error_code = 404
            return response, error_code
    except:
        response = jsonify({'result':'error','msg': f'template_id {template_id} inválido'})
        error_code = 400
        return response, error_code
    


# Endpoint para retornar link de topology graph / verify later
@app.route('/templates/<string:template_id>/graph', methods=['GET'])
def buscar_plantilla(template_id):
    ### ACCIÓN PARA RECOGER EL TEMPLATE USANDO EL TEMPLATE_ID. OBTENER SOLO EL OBJETO TOPOLOGY (SIN ESPECIFICACIONES)
    
    
    ###
    link = "link"
    response = jsonify({'result': 'success','topology_link': link, 'msg': f'URL de topologia de plantilla con template_id {template_id} obtenida correctamente'})

    
if __name__ == '__main__':
    app.run(debug=True)